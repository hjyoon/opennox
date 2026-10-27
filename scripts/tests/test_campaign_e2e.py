import importlib.util
import json
from pathlib import Path
import subprocess
import sys
import tempfile
from types import SimpleNamespace
import unittest
from unittest import mock

spec = importlib.util.spec_from_file_location(
    'campaign_e2e', Path(__file__).resolve().parents[1] / 'run-campaign-e2e.py')
runner = importlib.util.module_from_spec(spec)
spec.loader.exec_module(runner)


class SuiteFieldsTests(unittest.TestCase):
    def test_published_suites_keep_their_outcome_requirements(self):
        for path in (Path(runner.__file__).parent / 'e2e' / 'campaign').glob('*.json'):
            with self.subTest(path=path.name):
                runner.validate_suite_fields(json.loads(path.read_text()))

    def test_published_scenarios_only_replay_allowed_inputs(self):
        for path in (Path(runner.__file__).parent / 'e2e' / 'campaign').glob('*.json'):
            for case in json.loads(path.read_text())['cases']:
                with self.subTest(suite=path.name, case=case['name']):
                    runner.validate_scenario(path.parent / case['scenario'])

    def test_map_pickup_requires_an_unambiguous_original_item(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / 'pickup.yaml'
            step = {'action': 'pickup-map-object-input', 'extent': 1940, 'item': 'BluePotion'}

            def validate(value):
                path.write_text('steps:\n  - ' + json.dumps(value) + '\n  - {"action":"quit"}\n')
                return runner.validate_scenario(path)

            self.assertEqual(validate(step)[0], step)
            for extent in (None, 0, -1, 1 << 32, True, '1940'):
                with self.subTest(extent=extent), self.assertRaises(ValueError):
                    validate(dict(step, extent=extent))
            for item in (None, '', 638):
                with self.subTest(item=item), self.assertRaises(ValueError):
                    validate(dict(step, item=item))

    def test_misspelled_requirement_is_rejected_before_launch_or_output(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            suite = root / 'suite.json'
            suite.write_text(json.dumps({'version': 1, 'name': 'typo', 'cases': [
                {'name': 'east', 'scenario': 'never-launched.yaml',
                 'expected_maps': ['wiz02c'], 'required_monster_deaths': [1000195877]}]}))
            command = [sys.executable, runner.__file__, '--suite', str(suite)]
            for name in ('binary', 'data', 'seed', 'output'):
                command += ['--' + name, str(root / name)]
            result = subprocess.run(command, capture_output=True, text=True, timeout=10)
            self.assertEqual(result.returncode, 2)
            self.assertIn('unknown campaign case fields: required_monster_deaths', result.stderr)
            self.assertFalse((root / 'output').exists())


class RequiredMapPickupTests(unittest.TestCase):
    def setUp(self):
        self.player = {'ind': 0, 'class': ['PLAYER']}
        self.food = {'ind': 1939, 'type': {'id': 'Meat'}}
        self.required = [{'extent': 1939, 'item': 'Meat'}]
        self.log = ('NAMED PICKUP ACTUAL INPUT: id="Meat extent=1939" '
                    'wire=0x27a point=(522,402) cursor=2 frame=381\n')

    def check(self, samples, log=None):
        return runner.required_map_pickups(samples, self.log if log is None else log, self.required,
                                           actual_maps={'wiz02c'}, expected_maps={'wiz02c'})

    def test_original_item_actual_pickup_and_removal_are_all_required(self):
        self.assertTrue(self.check([[self.player, self.food], [self.player]])['passed'])
        for samples, log in (([[self.player], [self.player]], self.log),
                             ([[self.player, self.food], [self.player, self.food]], self.log),
                             ([[self.player, self.food], [self.player]], ''),
                             ([], self.log)):
            with self.subTest(samples=samples, log=log):
                self.assertFalse(self.check(samples, log)['passed'])

    def test_wrong_identity_or_cursor_does_not_prove_pickup(self):
        for log in (self.log.replace('1939', '1938'), self.log.replace('Meat', 'RedPotion'),
                    self.log.replace('cursor=2', 'cursor=20'), self.log.replace('cursor=2', 'cursor=1')):
            with self.subTest(log=log):
                self.assertFalse(self.check([[self.player, self.food], [self.player]], log)['passed'])
        wrong_type = dict(self.food, type={'id': 'RedPotion'})
        self.assertFalse(self.check([[self.player, wrong_type], [self.player]])['passed'])

    def test_unloaded_scene_cannot_prove_item_removal(self):
        result = self.check([[self.player, self.food], []])
        self.assertFalse(result['passed'])
        self.assertFalse(result['final_player_present'])

    def test_extent_reuse_in_another_map_cannot_prove_pickup(self):
        for maps in ({'wiz02a'}, {'wiz02a', 'wiz02c'}, set()):
            with self.subTest(maps=maps):
                result = runner.required_map_pickups(
                    [[self.player, self.food], [self.player]], self.log, self.required,
                    actual_maps=maps, expected_maps={'wiz02c'})
                self.assertFalse(result['passed'])
                self.assertFalse(result['map_scope_passed'])

    def test_missing_one_of_two_foods_fails_despite_healthy_player(self):
        other = {'ind': 1938, 'type': {'id': 'Meat'}}
        result = runner.required_map_pickups(
            [[self.player, self.food, other], [self.player, other]], self.log,
            self.required + [{'extent': 1938, 'item': 'Meat'}],
            actual_maps={'wiz02c'}, expected_maps={'wiz02c'})
        self.assertFalse(result['passed'])
        self.assertEqual([row['extent'] for row in result['objects'] if not row['passed']], [1938])

    def test_requirements_are_validated_before_launch(self):
        valid = {'cases': [{'expected_maps': ['wiz02c'], 'required_map_pickups': self.required}]}
        runner.validate_suite_fields(valid)
        for requirement in (None, [], [dict(extent=True, item='Meat')],
                            [dict(extent=-1, item='Meat')], [dict(extent=1, item='')],
                            [dict(extent=1, item='Meat', typo=True)], self.required * 2):
            with self.subTest(requirement=requirement), self.assertRaises(ValueError):
                runner.validate_suite_fields({'cases': [dict(expected_maps=['wiz02c'],
                                                           required_map_pickups=requirement)]})
        with self.assertRaises(ValueError):
            runner.validate_suite_fields({'cases': [dict(expected_maps=['wiz02b', 'wiz02c'],
                                                       required_map_pickups=self.required)]})


class PlayerOutcomeTests(unittest.TestCase):
    def test_failed_recovery_reports_actual_values_and_unchanged_minimums(self):
        # Preserved upper-cave recovery: a living manual Save is still below HP20.
        player = {'player': {'health': 17, 'maximum_health': 26},
                  'mana': 55, 'maximum_mana': 88}
        self.assertEqual(runner.player_value_failures(player, {
            'minimum_health': 20, 'minimum_mana': 52}),
            [{'check': 'minimum_health', 'minimum': 20, 'actual': 17}])

    def test_resource_minimums_keep_combat_variation_and_reject_missing_recovery(self):
        case = {'minimum_health': 20, 'minimum_mana': 52,
                'expected_maximum_health': 26, 'expected_maximum_mana': 88}
        for health in (20, 21, 22, 26):
            player = {'player': {'health': health, 'maximum_health': 26},
                      'mana': 54, 'maximum_mana': 88}
            self.assertTrue(runner.expected_player_values(player, case))
        for health, mana in ((19, 54), (22, 51), (22, 4), (None, 54), (22, None)):
            player = {'player': {'health': health, 'maximum_health': 26},
                      'mana': mana, 'maximum_mana': 88}
            self.assertFalse(runner.expected_player_values(player, case))
        self.assertFalse(runner.expected_player_values(
            {'player': {'health': 22}, 'mana': 54},
            {'minimum_health': 20, 'expected_health': 21}))

    def test_resource_minimums_are_validated_before_launch(self):
        for value in (-1, 0x10000, True, '20', None):
            with self.subTest(value=value), self.assertRaises(ValueError):
                runner.validate_suite_fields({'cases': [{'minimum_health': value}]})
        with self.assertRaises(ValueError):
            runner.validate_suite_fields({'cases': [{'minimum_mana': 52, 'expected_mana': 51}]})
        runner.validate_suite_fields({'cases': [{'minimum_health': 20, 'minimum_mana': 52}]})

    def test_earned_level_gain_requires_the_reported_maxima(self):
        # Actual final values from the original FlyingGolem2348 level gain.
        player = {'player': {'health': 26, 'maximum_health': 26},
                  'mana': 88, 'maximum_mana': 88}
        case = {'expected_health': 26, 'expected_mana': 88,
                'expected_maximum_health': 26, 'expected_maximum_mana': 88}
        self.assertTrue(runner.expected_player_values(player, case))
        for key, value in (('maximum_health', 27), ('maximum_mana', 89)):
            observed = dict(player, player=dict(player['player']))
            if key == 'maximum_health':
                observed['player'][key] = value
            else:
                observed[key] = value
            self.assertFalse(runner.expected_player_values(observed, case))

    def test_required_missing_values_fail_and_existing_suites_keep_current_checks(self):
        player = {'player': {'health': 20}, 'mana': 78}
        self.assertTrue(runner.expected_player_values(player, {'expected_health': 20, 'expected_mana': 78}))
        self.assertFalse(runner.expected_player_values(player, {'expected_maximum_health': 26}))
        self.assertFalse(runner.expected_player_values(player, {'expected_health': 26}))
        self.assertFalse(runner.expected_player_values(player, {'expected_maximum_mana': 88}))


class SequenceTests(unittest.TestCase):
    def test_published_sequences_validate_all_stages_without_gameplay(self):
        directory = Path(runner.__file__).parent / 'e2e' / 'campaign' / 'sequences'
        for path in directory.glob('*.json'):
            with self.subTest(path=path.name):
                _, stages = runner.load_sequence(path)
                self.assertTrue(stages)

    def prepare(self, root):
        binary = root / 'prebuilt'
        binary.write_bytes(b'one frozen executable')
        seed = root / 'earned' / 'Save'
        (seed / 'SAVE0001').mkdir(parents=True)
        (seed / 'SAVE0001' / 'Player.plr').write_bytes(b'original earned save')
        assets = root / 'assets'
        (assets / 'maps').mkdir(parents=True)
        for name in ('GAME.EXE', 'thing.bin'):
            (assets / name).write_bytes(name.encode())
        (root / 'route.yaml').write_text('steps:\n  - {"action":"quit"}\n')
        suite = {'version': 1, 'name': 'stage', 'cases': [
            {'name': 'case', 'scenario': 'route.yaml', 'expected_maps': ['wiz02c']}]}
        for name in ('first', 'second', 'third'):
            (root / (name + '.json')).write_text(json.dumps(suite))
        plan = root / 'sequence.json'
        plan.write_text(json.dumps({'version': 1, 'name': 'sequence', 'stages': [
            {'name': name, 'suite': name + '.json'} for name in ('first', 'second', 'third')]}))
        args = SimpleNamespace(sequence=plan, binary=binary, seed=seed, data=root / 'assets',
                               output=root / 'output', config_seed=None, sandbox_profile=None,
                               port=24973)
        return args

    def child(self, command, passed=True, **kwargs):
        options = dict(zip(command[2::2], command[3::2]))
        output = Path(options['--output'])
        runtime = output / 'case' / 'runtime-data.test'
        saved = runtime / 'Save' / 'SAVE0001' / 'Player.plr'
        saved.parent.mkdir(parents=True)
        saved.write_bytes(output.name.encode())
        suite_path = Path(options['--suite'])
        scenario = suite_path.parent / json.loads(suite_path.read_text())['cases'][0]['scenario']
        preserved = output / 'case' / 'scenario.yaml'
        preserved.write_bytes(scenario.read_bytes())
        (output / 'suite.source.json').write_bytes(suite_path.read_bytes())
        (output / 'opennox-tested').write_bytes(Path(options['--binary']).read_bytes())
        routes = {}
        for name, route in runner.ROUTES.items():
            path = output / 'case' / (name + '-0.json')
            path.write_text('[]')
            routes[name] = {'method': 'GET', 'route': route, 'observation_only': True,
                            'status': 200, 'path': str(path), 'bytes': path.stat().st_size,
                            'sha256': runner.sha(path)}
        report = {'passed': passed, 'binary_sha256': runner.sha(Path(options['--binary'])),
                  'source_seed_manifest': runner.tree(Path(options['--seed'])),
                  'seed_unchanged': True, 'data_inputs_unchanged': True,
                  'config_unchanged': True, 'binary_unchanged': True,
                  'cases': [{'passed': passed, 'continuation_eligible': True,
                             'manual_save_written': True, 'runtime_data': str(runtime),
                             'saved_player_sha256': runner.sha(saved),
                             'scenario': str(preserved), 'scenario_sha256': runner.sha(scenario),
                             'snapshots': [{'routes': routes}]}]}
        runner.put(output / 'result.json', report)
        data = Path(options.get('--data', str(suite_path.parent / 'assets')))
        runner.put(output / 'inputs.json', {
            'suite_sha256': runner.sha(suite_path), 'binary_sha256': report['binary_sha256'],
            'seed': report['source_seed_manifest'],
            'config_sha256': runner.sha(Path(options['--config-seed'])) if '--config-seed' in options else None,
            'data': {str(data / name): runner.sha(data / name) for name in ('GAME.EXE', 'thing.bin')}})
        # Also test exit 0 with failed objectives: healthy alone must never advance.
        return SimpleNamespace(returncode=0)

    def test_sequence_passes_exact_saves_forward_with_one_binary_and_no_rebuild(self):
        with tempfile.TemporaryDirectory() as directory:
            args = self.prepare(Path(directory))
            sequence, stages = runner.load_sequence(args.sequence)
            with mock.patch.object(runner.subprocess, 'run', side_effect=self.child) as launch, mock.patch('builtins.print'):
                self.assertEqual(runner.run_sequence(args, sequence, stages), 0)
            commands = [call.args[0] for call in launch.call_args_list]
            options = [dict(zip(cmd[2::2], cmd[3::2])) for cmd in commands]
            self.assertEqual(len({row['--binary'] for row in options}), 1)
            self.assertEqual(options[0]['--seed'], str(args.seed))
            for previous, following in zip(options, options[1:]):
                self.assertEqual(following['--seed'], str((Path(previous['--output']) / 'case/runtime-data.test/Save').resolve()))
            report = json.loads((args.output / 'result.json').read_text())
            self.assertTrue(report['passed'])
            self.assertEqual(report['final_seed_manifest'], runner.tree(Path(report['final_seed'])))
            self.assertEqual((args.seed / 'SAVE0001/Player.plr').read_bytes(), b'original earned save')
            self.assertFalse(report['overall_porting_100_percent'])

    def test_failed_objective_stops_sequence_even_with_healthy_save_and_exit_zero(self):
        with tempfile.TemporaryDirectory() as directory:
            args = self.prepare(Path(directory))
            sequence, stages = runner.load_sequence(args.sequence)
            def launch(command, **kwargs):
                output = Path(command[command.index('--output') + 1])
                result = self.child(command, passed=output.name != 'second')
                if output.name == 'second':
                    path = output / 'result.json'
                    child = json.loads(path.read_text())
                    child['cases'][0].update(name='upper-urchins', failed_checks=[{
                        'check': 'required_defeated_script_ids', 'missing': [1000194053],
                        'initial_alive': [1000194053, 1000194055], 'observed_dead': [1000194055]}])
                    path.write_text(json.dumps(child))
                return result
            with mock.patch.object(runner.subprocess, 'run', side_effect=launch) as run, mock.patch('builtins.print'):
                self.assertEqual(runner.run_sequence(args, sequence, stages), 1)
            self.assertEqual(run.call_count, 2)
            report = json.loads((args.output / 'result.json').read_text())
            self.assertFalse(report['passed'])
            self.assertEqual(report['stages'][-1]['failed_checks'], [{
                'case': 'upper-urchins', 'check': 'required_defeated_script_ids',
                'missing': [1000194053], 'initial_alive': [1000194053, 1000194055],
                'observed_dead': [1000194055]}])
            self.assertNotIn('final_seed', report)
            self.assertFalse((args.output / 'third').exists())
            self.assertTrue((args.output / 'second/case/runtime-data.test/Save/SAVE0001/Player.plr').exists())

    def test_later_invalid_stage_is_rejected_before_any_game_launch(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            args = self.prepare(root)
            suite = json.loads((root / 'second.json').read_text())
            suite['cases'][0]['required_monster_deaths'] = [1000195877]
            (root / 'second.json').write_text(json.dumps(suite))
            with self.assertRaisesRegex(ValueError, 'unknown campaign case fields'):
                runner.load_sequence(args.sequence)
            self.assertFalse(args.output.exists())

    def test_continuation_rejects_wrong_binary_seed_changed_save_and_external_runtime(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            args = self.prepare(root)
            stage = root / 'stage'
            command = ['python', 'runner', '--suite', str(root / 'first.json'),
                       '--binary', str(args.binary), '--seed', str(args.seed), '--output', str(stage)]
            self.child(command)
            original = json.loads((stage / 'result.json').read_text())
            expected_seed, expected_binary = runner.tree(args.seed), runner.sha(args.binary)
            for change in ({'binary_sha256': 'different'}, {'source_seed_manifest': {}},
                           {'passed': False}):
                with self.subTest(change=change), self.assertRaises(ValueError):
                    runner.continuation_seed(dict(original, **change), stage, {}, expected_binary, expected_seed)
            changed = json.loads(json.dumps(original))
            changed['cases'][0]['runtime_data'] = str(args.seed.parent)
            with self.assertRaisesRegex(ValueError, 'inside its report directory'):
                runner.continuation_seed(changed, stage, {}, expected_binary, expected_seed)
            saved = Path(original['cases'][0]['runtime_data']) / 'Save/SAVE0001/Player.plr'
            saved.write_bytes(b'changed after validation')
            with self.assertRaisesRegex(ValueError, 'has changed'):
                runner.continuation_seed(original, stage, {}, expected_binary, expected_seed)


class SequenceResumeTests(unittest.TestCase):
    prepare = SequenceTests.prepare
    child = SequenceTests.child

    def previous(self, root, failed='second'):
        args = self.prepare(root)
        sequence, stages = runner.load_sequence(args.sequence)

        def launch(command, **kwargs):
            output = Path(command[command.index('--output') + 1])
            return self.child(command, passed=output.name != failed)

        with mock.patch.object(runner.subprocess, 'run', side_effect=launch), mock.patch('builtins.print'):
            runner.run_sequence(args, sequence, stages)
        args.resume_from = args.output
        args.output = root / 'resumed'
        return args

    def test_replays_only_failed_and_later_stages_and_keeps_failure_evidence(self):
        with tempfile.TemporaryDirectory() as directory:
            args = self.previous(Path(directory))
            before = runner.tree(args.resume_from)
            sequence, stages = runner.load_sequence(args.sequence)
            with mock.patch.object(runner.subprocess, 'run', side_effect=self.child) as launch, mock.patch('builtins.print'):
                self.assertEqual(runner.run_sequence(args, sequence, stages), 0)
            commands = [call.args[0] for call in launch.call_args_list]
            self.assertEqual([Path(cmd[cmd.index('--output') + 1]).name for cmd in commands], ['second', 'third'])
            report = json.loads((args.output / 'result.json').read_text())
            self.assertEqual(report['reused_stages'], 1)
            self.assertEqual(report['executed_stages'], 2)
            self.assertFalse(report['fresh_replay'])
            self.assertTrue(report['stages'][0]['reused'])
            self.assertEqual(commands[0][commands[0].index('--seed') + 1], report['stages'][0]['next_seed'])
            self.assertEqual(runner.tree(args.resume_from), before)

    def test_completed_replay_can_be_reused_again_without_game_launch(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            args = self.previous(root, failed=None)
            sequence, stages = runner.load_sequence(args.sequence)
            for destination in ('resumed', 'resumed-again'):
                args.output = root / destination
                with mock.patch.object(runner.subprocess, 'run') as launch, mock.patch('builtins.print'):
                    self.assertEqual(runner.run_sequence(args, sequence, stages), 0)
                launch.assert_not_called()
                report = json.loads((args.output / 'result.json').read_text())
                self.assertEqual(report['reused_stages'], 3)
                self.assertEqual(report['executed_stages'], 0)
                self.assertFalse(report['fresh_replay'])
                args.resume_from = args.output

    def test_changed_binary_inputs_save_or_evidence_reject_before_launch(self):
        changes = ('binary', 'seed', 'suite', 'scenario', 'data', 'config', 'player-save',
                   'map-save', 'snapshot', 'preserved-scenario', 'preserved-suite')
        for change in changes:
            with self.subTest(change=change), tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                args = self.previous(root)
                stage_dir = args.resume_from / 'first'
                child = json.loads((stage_dir / 'result.json').read_text())['cases'][0]
                paths = {'binary': args.binary, 'seed': args.seed / 'SAVE0001/Player.plr',
                         'suite': root / 'first.json', 'scenario': root / 'route.yaml',
                         'data': args.data / 'thing.bin',
                         'player-save': Path(child['runtime_data']) / 'Save/SAVE0001/Player.plr',
                         'map-save': Path(child['runtime_data']) / 'Save/SAVE0001/changed.map',
                         'snapshot': Path(child['snapshots'][0]['routes']['objects']['path']),
                         'preserved-scenario': Path(child['scenario']),
                         'preserved-suite': stage_dir / 'suite.source.json'}
                if change == 'config':
                    args.config_seed = root / 'different.cfg'
                    args.config_seed.write_text('different config')
                else:
                    path = paths[change]
                    path.write_bytes(path.read_bytes() + b'\n' if path.exists() else b'changed map')
                sequence, stages = runner.load_sequence(args.sequence)
                with mock.patch.object(runner.subprocess, 'run') as launch, self.assertRaises(ValueError):
                    runner.run_sequence(args, sequence, stages)
                launch.assert_not_called()
                self.assertFalse(args.output.exists())

    def test_failed_stage_route_can_change_without_replaying_passed_prefix(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            args = self.previous(root)
            suite = json.loads((root / 'second.json').read_text())
            suite['cases'][0]['scenario'] = 'new-route.yaml'
            (root / 'new-route.yaml').write_text('steps:\n  - {"action":"wait","dt":1}\n  - {"action":"quit"}\n')
            (root / 'second.json').write_text(json.dumps(suite))
            sequence, stages = runner.load_sequence(args.sequence)
            with mock.patch.object(runner.subprocess, 'run', side_effect=self.child) as launch, mock.patch('builtins.print'):
                self.assertEqual(runner.run_sequence(args, sequence, stages), 0)
            self.assertEqual(launch.call_count, 2)

    def test_interrupted_sequence_uses_only_complete_passing_progress_records(self):
        with tempfile.TemporaryDirectory() as directory:
            args = self.previous(Path(directory))
            (args.resume_from / 'result.json').unlink()
            progress = args.resume_from / 'stages.jsonl'
            first = progress.read_text().splitlines()[0]
            progress.write_text(first + '\n{"name":"second",')
            sequence, stages = runner.load_sequence(args.sequence)
            with mock.patch.object(runner.subprocess, 'run', side_effect=self.child) as launch, mock.patch('builtins.print'):
                self.assertEqual(runner.run_sequence(args, sequence, stages), 0)
            self.assertEqual(launch.call_count, 2)

    def test_cli_rejects_resume_for_single_suite_and_output_inside_old_report(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            args = self.previous(root)
            for selection, output in (('--suite', root / 'single-output'),
                                      ('--sequence', args.resume_from / 'new-output')):
                source = root / 'first.json' if selection == '--suite' else args.sequence
                command = [sys.executable, runner.__file__, selection, str(source),
                           '--resume-from', str(args.resume_from)]
                for name in ('binary', 'data', 'seed'):
                    command += ['--' + name, str(getattr(args, name))]
                command += ['--output', str(output)]
                result = subprocess.run(command, capture_output=True, text=True, timeout=10)
                self.assertEqual(result.returncode, 2)
                self.assertFalse(output.exists())
                self.assertIn('resume', result.stderr)


class ScenarioEvidenceTests(unittest.TestCase):
    def test_launch_uses_preserved_input_when_the_source_is_edited(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            source = root / 'route.yaml'
            original = 'steps:\n  - {"action":"quit"}\n'
            source.write_text(original)
            output = root / 'output'
            output.mkdir()
            args = SimpleNamespace(output=output, seed=root / 'Save', data=root,
                                   config_seed=None, sandbox_profile=None)
            launch = subprocess.Popen

            def edit_source_at_launch(command, **kwargs):
                source.write_text('the next route has replaced the source\n')
                self.assertEqual(Path(command[-2]), output / 'case' / 'scenario.yaml')
                self.assertEqual(Path(command[-2]).read_text(), original)
                return launch([sys.executable, '-c', 'pass'], **kwargs)

            with mock.patch.object(runner, 'check_ports'), mock.patch.object(
                    runner.subprocess, 'Popen', side_effect=edit_source_at_launch), mock.patch('builtins.print'):
                report = runner.run_case(args, root / 'prebuilt', root,
                                         {'name': 'case', 'scenario': source.name,
                                          'expected_maps': ['wiz02c']}, 24989)
            self.assertEqual(report['source_scenario'], str(source.resolve()))
            self.assertEqual(Path(report['scenario']).read_text(), original)
            self.assertEqual(report['scenario_sha256'], runner.sha(output / 'case' / 'scenario.yaml'))
            # A preservation test without gameplay evidence cannot certify a case.
            self.assertFalse(report['passed'])


def spider(ident, dead=False):
    return {'script_id': ident, 'type': {'id': 'Spider'},
            'flags': ['DEAD'] if dead else ['ACTIVE']}


class CreatedMonsterHistoryTests(unittest.TestCase):
    def test_observed_deaths_survive_corpse_cleanup(self):
        # Reduced from the actual Wiz02c first-scene capture: both summoned
        # Spiders were DEAD before natural recovery removed their corpses.
        samples = [[], [spider(1000198888), spider(1000198889)],
                   [spider(1000198888, True), spider(1000198889, True)], [], []]
        history = runner.created_monster_history(samples, 'Spider')
        self.assertEqual(history['count'], 2)
        self.assertEqual(history['dead'], 2)
        self.assertEqual(history['observed_dead_script_ids'], [1000198888, 1000198889])
        self.assertEqual(history['absent_from_final_script_ids'], [1000198888, 1000198889])

    def test_disappearance_without_observed_death_is_not_a_kill(self):
        history = runner.created_monster_history([[], [spider(10)], []], 'Spider')
        self.assertEqual(history['count'], 1)
        self.assertEqual(history['dead'], 0)

    def test_existing_actors_and_repeated_samples_do_not_add_kills(self):
        samples = [[spider(1, True)], [spider(1, True), spider(2)],
                   [spider(1, True), spider(2, True)], [spider(2, True)]]
        history = runner.created_monster_history(samples, 'Spider')
        self.assertEqual(history['script_ids'], [2])
        self.assertEqual(history['count'], 1)
        self.assertEqual(history['dead'], 1)

    def test_later_resurrection_cancels_the_death_claim(self):
        samples = [[], [spider(2, True)], [spider(2)], []]
        history = runner.created_monster_history(samples, 'Spider')
        self.assertEqual(history['count'], 1)
        self.assertEqual(history['observed_dead_script_ids'], [2])
        self.assertEqual(history['dead'], 0)


class RequiredMonsterDeathsTests(unittest.TestCase):
    @staticmethod
    def monster(ident, dead=False):
        return dict(spider(ident, dead), **{'class': ['MONSTER']})

    def test_required_original_death_survives_corpse_cleanup(self):
        history = runner.required_monster_deaths(
            [[self.monster(2350)], [self.monster(2350, True)], []], [2350])
        self.assertTrue(history['passed'])
        self.assertEqual(history['last_observed_dead_script_ids'], [2350])

    def test_preexisting_dead_missing_and_later_created_do_not_count(self):
        history = runner.required_monster_deaths(
            [[self.monster(1, True)], [self.monster(1, True), self.monster(2, True)]],
            [1, 2, 3])
        self.assertFalse(history['passed'])
        self.assertEqual(history['initial_alive_script_ids'], [])

    def test_disappearance_without_death_does_not_count(self):
        history = runner.required_monster_deaths([[self.monster(1)], []], [1])
        self.assertFalse(history['passed'])

    def test_later_resurrection_cancels_required_death(self):
        history = runner.required_monster_deaths(
            [[self.monster(1)], [self.monster(1, True)], [self.monster(1)], []], [1])
        self.assertFalse(history['passed'])
        self.assertEqual(history['observed_dead_script_ids'], [1])
        self.assertEqual(history['last_observed_dead_script_ids'], [])


class RequiredEnableTransitionsTests(unittest.TestCase):
    requirements = [{'script_id': 1985, 'before': True, 'after': False},
                    {'script_id': 1982, 'before': False, 'after': True}]

    @staticmethod
    def sample(plate=True, mover=False):
        return [{'script_id': 1985, 'enabled': plate},
                {'script_id': 1982, 'enabled': mover}]

    def test_original_plate_disables_and_mover_enables(self):
        # Reduced from the original Bookcase6 normal/HD snapshots.
        result = runner.required_enable_transitions(
            [self.sample(), self.sample(False, True), self.sample(False, True)],
            self.requirements)
        self.assertTrue(result['passed'])

    def test_previously_activated_switch_does_not_count(self):
        self.assertFalse(runner.required_enable_transitions(
            [self.sample(False, True), self.sample(False, True)], self.requirements)['passed'])

    def test_missing_or_later_created_actor_does_not_count(self):
        for samples in ([], [[], self.sample(False, True)], [self.sample(), []]):
            with self.subTest(samples=samples):
                self.assertFalse(runner.required_enable_transitions(samples, self.requirements)['passed'])

    def test_later_reversion_cancels_the_transition(self):
        self.assertFalse(runner.required_enable_transitions(
            [self.sample(), self.sample(False, True), self.sample()], self.requirements)['passed'])

    def test_boolean_states_and_unique_identity_are_required(self):
        for samples in ([self.sample(1, 0), self.sample(0, 1)],
                        [self.sample(), self.sample(False, True) + [{'script_id': 1982, 'enabled': True}]]):
            with self.subTest(samples=samples):
                self.assertFalse(runner.required_enable_transitions(samples, self.requirements)['passed'])
        self.assertFalse(runner.required_enable_transitions(
            [[{'script_id': 1}], [{'script_id': 1}]],
            [{'script_id': 1, 'before': None, 'after': None}])['passed'])


class RequiredPositionTransitionsTests(unittest.TestCase):
    requirements = [
        {'script_id': 1000193761, 'before': {'x': 1957, 'y': 3780}, 'after': {'x': 1871, 'y': 3998}},
        {'script_id': 1000193759, 'before': {'x': 1618, 'y': 4109}, 'after': {'x': 1818, 'y': 3928}},
    ]

    def sample(self, first=False, second=False):
        return [{'script_id': row['script_id'], 'pos': row['after' if moved else 'before']}
                for row, moved in zip(self.requirements, (first, second))]

    def test_original_switches_move_both_lights(self):
        samples = [self.sample(), self.sample(True), self.sample(True, True)]
        self.assertTrue(runner.required_position_transitions(samples, self.requirements)['passed'])

    def test_prepositioned_or_unmoved_lights_do_not_count(self):
        for samples in ([self.sample(True, True), self.sample(True, True)],
                        [self.sample(), self.sample(True)]):
            with self.subTest(samples=samples):
                self.assertFalse(runner.required_position_transitions(samples, self.requirements)['passed'])

    def test_missing_later_created_or_duplicate_identity_fails(self):
        for samples in ([], [[], self.sample(True, True)], [self.sample(), []],
                        [self.sample(), self.sample(True, True) * 2, self.sample(True, True)]):
            with self.subTest(samples=samples):
                self.assertFalse(runner.required_position_transitions(samples, self.requirements)['passed'])

    def test_later_reversion_cancels_the_transition(self):
        samples = [self.sample(), self.sample(True, True), self.sample()]
        self.assertFalse(runner.required_position_transitions(samples, self.requirements)['passed'])

    def test_nonfinite_coordinates_and_invalid_or_overlapping_bounds_fail(self):
        for bad in (float('nan'), float('inf'), True, None):
            samples = [self.sample(), self.sample(True, True)]
            samples[-1][0] = dict(samples[-1][0], pos={'x': bad, 'y': 3998})
            self.assertFalse(runner.required_position_transitions(samples, self.requirements)['passed'])
        for changes in ({'tolerance': -1}, {'tolerance': float('nan')}, {'tolerance': True},
                        {'tolerance': 200}, {'before': {'x': None, 'y': 3780}}):
            required = [dict(self.requirements[0], **changes)]
            self.assertFalse(runner.required_position_transitions(
                [self.sample(), self.sample(True, True)], required)['passed'])


class RequiredPlateHoldsTests(unittest.TestCase):
    requirements = [{'weight_script_id': 2400, 'plate_script_id': 2399,
                     'player_script_id': 1, 'maximum_weight_distance': 16,
                     'minimum_player_distance': 175, 'maximum_weight_drift': .5}]

    @staticmethod
    def sample(weight=(0, -101), player=(0, -200)):
        return [{'script_id': ident, 'pos': {'x': xy[0], 'y': xy[1]}}
                for ident, xy in ((2400, weight), (2399, (0, 0)), (1, player))]

    def test_crate_stays_on_plate_as_player_leaves(self):
        samples = [self.sample(), self.sample((-12, -4), (0, -30)),
                   self.sample((-12, -4), (180, 0)), self.sample((-12, -4), (400, 0))]
        self.assertTrue(runner.required_plate_holds(samples, self.requirements)['passed'])

    def test_prepositioned_or_absent_weight_does_not_count(self):
        for samples in ([], [self.sample((0, 0)), self.sample((0, 0))],
                        [self.sample(), self.sample()[1:]],
                        [self.sample()[1:], self.sample((0, 0)), self.sample((0, 0))]):
            with self.subTest(samples=samples):
                self.assertFalse(runner.required_plate_holds(samples, self.requirements)['passed'])

    def test_player_standing_on_plate_does_not_prove_weight_hold(self):
        self.assertFalse(runner.required_plate_holds(
            [self.sample(), self.sample((0, 0), (0, 0)), self.sample((0, 0), (0, 0))],
            self.requirements)['passed'])

    def test_weight_leaving_or_drifting_after_separation_fails(self):
        for middle in ((0, 40), (10, 0)):
            with self.subTest(middle=middle):
                samples = [self.sample(), self.sample((0, 0)), self.sample(middle), self.sample((0, 0))]
                self.assertFalse(runner.required_plate_holds(samples, self.requirements)['passed'])

    def test_one_observation_nonfinite_and_duplicate_identity_fail(self):
        for samples in ([self.sample(), self.sample((0, 0))],
                        [self.sample(), self.sample((0, 0)), self.sample((float('nan'), 0))],
                        [self.sample(), self.sample((0, 0)), self.sample((0, 0)) + self.sample((0, 0))[:1]]):
            with self.subTest(samples=samples):
                self.assertFalse(runner.required_plate_holds(samples, self.requirements)['passed'])


if __name__ == '__main__':
    unittest.main()
