#!/usr/bin/env python3
"""Portable read-only access to the sealed DefaultDamage original reference ZIP.

Absolute paths inside authentic records remain unchanged. resolve_path maps
them through the index; snapshot refs resolve to the shared memory pool.
"""
import argparse
from collections import OrderedDict
import gzip
import hashlib
import io
import json
from pathlib import Path
import zipfile


class ReferenceBundle:
    def __init__(self, path, cache_bytes=16 << 20):
        self.zip = zipfile.ZipFile(path, 'r')
        self.index = json.loads(self.zip.read('bundle-index-v1.json'))
        assert self.index['schema'] == 'opennox-default-damage-parent-portable-reference-bundle-v1'
        self.members = self.index['members']
        assert set(self.zip.namelist()) == set(self.members) | {'bundle-index-v1.json'}
        assert len(self.zip.namelist()) == len(set(self.zip.namelist()))
        self.paths = {path: name for name, row in self.members.items() for path in row['canonical_source_paths']}
        self.limit, self.cached_bytes, self.cache = cache_bytes, 0, OrderedDict()
        self.raw_verified = set()

    def close(self):
        self.cache.clear()
        self.zip.close()

    def resolve_path(self, canonical_old_path):
        """No filesystem fallback and no rewriting of authentic source records."""
        return self.paths[canonical_old_path]

    def read(self, member):
        row = self.members[member]
        data = self.zip.read(member)
        assert len(data) == row['bytes'] and hashlib.sha256(data).hexdigest() == row['sha256']
        return data

    def json(self, member):
        return json.loads(self.read(member))

    def run(self, name):
        return self.index['runs'][name]

    def rows(self, name):
        member = self.run(name)['rows_member']
        expected = self.members[member]
        hasher, count = hashlib.sha256(), 0
        with self.zip.open(member) as raw:
            while chunk := raw.read(1 << 20):
                hasher.update(chunk)
                count += len(chunk)
        assert (count, hasher.hexdigest()) == (expected['bytes'], expected['sha256'])
        with self.zip.open(member) as packed:
            with gzip.GzipFile(fileobj=packed) as stream:
                for line in stream:
                    assert line.endswith(b'\n')
                    yield json.loads(line)

    def snapshot(self, ref):
        member = ref['file']
        assert member == 'memory/'+ref['sha256']+'.bin.gz'
        if member in self.cache:
            result = self.cache.pop(member)
            self.cache[member] = result
            return result
        compressed = self.read(member)
        assert len(compressed) == ref['gzip_bytes'] and hashlib.sha256(compressed).hexdigest() == ref['gzip_sha256']
        data = gzip.decompress(compressed)
        assert len(data) == ref['bytes'] and hashlib.sha256(data).hexdigest() == ref['sha256']
        while self.cache and self.cached_bytes+len(data) > self.limit:
            _, removed = self.cache.popitem(last=False)
            self.cached_bytes -= len(removed)
        if len(data) <= self.limit:
            self.cache[member] = data
            self.cached_bytes += len(data)
        return data

    def verify_all(self):
        total = decoded = 0
        for name, expected in self.members.items():
            hasher, count = hashlib.sha256(), 0
            with self.zip.open(name) as f:
                while chunk := f.read(1 << 20):
                    hasher.update(chunk)
                    count += len(chunk)
            assert (count, hasher.hexdigest()) == (expected['bytes'], expected['sha256']), name
            total += count
            if 'uncompressed_memory' in expected:
                memory = expected['uncompressed_memory']
                hasher, count = hashlib.sha256(), 0
                with self.zip.open(name) as raw:
                    with gzip.GzipFile(fileobj=raw) as stream:
                        while chunk := stream.read(1 << 20):
                            hasher.update(chunk)
                            count += len(chunk)
                assert (count, hasher.hexdigest()) == (memory['bytes'], memory['sha256']), name
                decoded += count
        return dict(members=len(self.members), member_bytes=total, decoded_memory_bytes=decoded,
                    all_raw_member_SHA_verified=True, all_full_memory_SHA_verified=True,
                    original_or_Go_or_game_execution=False)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('bundle', type=Path)
    parser.add_argument('--verify-all', action='store_true')
    args = parser.parse_args()
    bundle = ReferenceBundle(args.bundle)
    try:
        result = bundle.verify_all() if args.verify_all else dict(
            scope=bundle.index['scope'], runs=bundle.index['runs'], shared_memory=bundle.index['shared_memory'])
        print(json.dumps(result, indent=2))
    finally:
        bundle.close()


if __name__ == '__main__':
    main()
