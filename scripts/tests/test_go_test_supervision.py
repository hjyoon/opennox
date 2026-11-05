import importlib.util
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest


SCRIPTS = Path(__file__).resolve().parents[1]
WRAPPER = SCRIPTS / "run-preserved-go-test.py"


def load_collector():
    spec = importlib.util.spec_from_file_location("test_binary_collector", SCRIPTS / "preserve-go-test-binaries.py")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class TestPreExecutionPreservation(unittest.TestCase):
    def run_child(self, root, code, *, timeout=10, arguments=()):
        report = root / "report with spaces"
        command = [sys.executable, str(WRAPPER), "--report", str(report),
                   "--timeout-seconds", str(timeout), "--heartbeat-seconds", "0.1", "--",
                   sys.executable, "-c", code, str(report), *arguments]
        result = subprocess.run(command, cwd=root, stdout=subprocess.PIPE, stderr=subprocess.PIPE, timeout=15)
        return report, result

    def test_copy_exists_before_start_and_exit_streams_arguments_are_exact(self):
        code = r'''
import hashlib,json,os,sys
from pathlib import Path
report=Path(sys.argv[1]);record=json.loads(next((report/'test-binary-records').glob('*.json')).read_text())
assert record['preservation_before_execution']
retained=Path(record['retained_path'])
with retained.open('rb')as f:assert hashlib.file_digest(f,'sha256').hexdigest()==record['sha256']
(report/'observed-child.json').write_text(json.dumps(dict(arguments=sys.argv[2:],cwd=os.getcwd())),encoding='utf-8')
sys.stdout.buffer.write(b'actual child stdout\n');sys.stdout.buffer.flush()
sys.stderr.buffer.write(b'actual child stderr\n');sys.stderr.buffer.flush()
raise SystemExit(17)
'''
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            report, result = self.run_child(root, code, arguments=("argument with spaces", "한글"))
            self.assertEqual(result.returncode, 17, result.stderr)
            self.assertEqual(result.stdout, b"actual child stdout\n")
            self.assertEqual(result.stderr, b"actual child stderr\n")
            observed = json.loads((report / "observed-child.json").read_text())
            self.assertEqual(observed["arguments"], ["argument with spaces", "한글"])
            self.assertEqual(Path(observed["cwd"]).resolve(), root.resolve())
            record = json.loads(next((report / "test-binary-records").glob("*.json")).read_text())
            self.assertEqual(record["phase"], "exited")
            self.assertEqual(record["native_exit_code"], 17)
            self.assertTrue(record["execution_started"])
            self.assertFalse(record["forced_termination"])

    def test_independent_deadline_kills_owned_descendants_and_keeps_binary(self):
        code = r'''
import subprocess,sys,time
subprocess.Popen([sys.executable,'-c','import time; time.sleep(60)'])
print('test entered',flush=True)
time.sleep(60)
'''
        with tempfile.TemporaryDirectory() as directory:
            report, result = self.run_child(Path(directory), code, timeout=1)
            self.assertEqual(result.returncode, 124, result.stderr)
            self.assertEqual(result.stdout, b"test entered\n")
            record = json.loads(next((report / "test-binary-records").glob("*.json")).read_text())
            self.assertEqual(record["phase"], "timed-out")
            self.assertTrue(record["forced_termination"])
            self.assertNotEqual(record["native_exit_code"], 0)
            self.assertLess(record["elapsed_seconds"], 10)
            self.assertTrue(Path(record["retained_path"]).is_file())

    def test_changed_retained_copy_is_never_replaced_or_executed(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            report, first = self.run_child(root, "pass")
            self.assertEqual(first.returncode, 0, first.stderr)
            retained = report / "test-binaries" / Path(sys.executable).name
            retained.write_bytes(b"preserved damaged evidence")
            marker = root / "unexpected-execution"
            code = "from pathlib import Path; Path(" + repr(str(marker)) + ").touch()"
            _, second = self.run_child(root, code)
            self.assertNotEqual(second.returncode, 0)
            self.assertIn(b"Refusing to replace", second.stderr)
            self.assertFalse(marker.exists())
            self.assertEqual(retained.read_bytes(), b"preserved damaged evidence")

    def test_windows_exception_dword_is_preserved(self):
        if os.name != "nt":
            self.skipTest("native Windows process exit convention")
        with tempfile.TemporaryDirectory() as directory:
            report, result = self.run_child(Path(directory), "import ctypes; ctypes.windll.kernel32.ExitProcess(0xC0000005)")
            self.assertEqual(result.returncode, 0xC0000005)
            record = json.loads(next((report / "test-binary-records").glob("*.json")).read_text())
            self.assertEqual(record["native_exit_code"], 0xC0000005)
            self.assertFalse(record["forced_termination"])


class TestFinalCollection(unittest.TestCase):
    def fixture(self, root, count):
        report, work, output = root / "report", root / "work", root / "output"
        for path in (report, work, output):
            path.mkdir()
        (report / "test.stderr.log").write_text("")
        (report / "test.jsonl").write_text(json.dumps({"Action": "build-output", "Output": "WORK=" + str(work) + "\n"}) + "\n")
        for i, name in enumerate(["v1.test.exe", "server.test.exe", "legacy.test.exe"][:count]):
            directory = work / ("b" + str(i))
            directory.mkdir()
            (directory / name).write_bytes(b"built but not installed " + name.encode())
        return report, work, output

    def test_cancelled_output_directory_falls_back_to_original_work_binaries(self):
        with tempfile.TemporaryDirectory() as directory:
            report, _, output = self.fixture(Path(directory), 3)
            collector = load_collector()
            rows = collector.preserve(report, output)
            self.assertEqual(len(rows), 3)
            self.assertEqual(collector.preserve(report, output), rows)
            status = json.loads((report / "test-preservation-status.json").read_text())
            self.assertTrue(status["complete_package_executable_set"])
            self.assertEqual(status["pre_execution_records"], 0)

    def test_incomplete_cleanup_records_missing_binaries_without_claiming_completion(self):
        with tempfile.TemporaryDirectory() as directory:
            report, _, output = self.fixture(Path(directory), 1)
            collector = load_collector()
            rows = collector.preserve(report, output, allow_incomplete=True)
            self.assertEqual(len(rows), 1)
            status = json.loads((report / "test-preservation-status.json").read_text())
            self.assertFalse(status["complete_package_executable_set"])
            self.assertTrue(status["problems"])
            with self.assertRaises(ValueError):
                collector.preserve(report, output)


if __name__ == "__main__":
    unittest.main()
