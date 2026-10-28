import copy
import importlib.util
import json
import math
from pathlib import Path
import unittest

spec = importlib.util.spec_from_file_location('campaign_swap', Path(__file__).resolve().parents[1] / 'run-campaign-e2e.py')
runner = importlib.util.module_from_spec(spec)
spec.loader.exec_module(runner)


class NativeSwapEvidenceTests(unittest.TestCase):
    def setUp(self):
        self.required = [{'player_script_id': 1, 'target_script_ids': [2, 3], 'key': 33,
                          'player_start': [0, 0], 'minimum_distance': 25, 'tolerance': 4}]
        self.samples = [[{'script_id': 1, 'pos': {'x': 0, 'y': 0}},
                         {'script_id': 2, 'flags': ['ENABLED'], 'pos': {'x': 60, 'y': 30}},
                         {'script_id': 3, 'flags': ['ENABLED'], 'pos': {'x': -60, 'y': 30}}],
                        [{'script_id': 1, 'pos': {'x': 42, 'y': 22}}]]
        self.rows = [dict(frame=100, cast_frame=98, player_script_id=1, script_id=2,
                          extent=17, wire=19, key=33, key_queued=True, player=[0, 0], target=[40, 20]),
                     dict(frame=101, cast_frame=98, player_script_id=1, script_id=2,
                          extent=17, wire=19, key=33, key_queued=True, player=[42, 22], target=[1, 2])]
        self.input = 'MONSTER QUICKBAR ACTUAL INPUT: type="SmallAlbinoSpider" wire=0x13 extent=17 script_id=2 key=33 key_queued=true'

    def log(self, rows=None, actual=None):
        return '\n'.join(['MONSTER QUICKBAR POSITION: ' + json.dumps(r) for r in (self.rows if rows is None else rows)]
                         + [self.input if actual is None else actual])

    def check(self, rows=None, samples=None, actual=None):
        return runner.required_swap_pairs(self.samples if samples is None else samples,
                                          self.log(rows, actual), self.required)

    def test_actual_input_and_reciprocal_single_frame_exchange_accept_a_moving_target(self):
        # The target has already moved away from its first GET position.
        result = self.check()
        self.assertTrue(result['passed'])
        self.assertEqual(result['pairs'][0]['position_exchange']['after']['frame'], 101)

    def moving_tail(self):
        rows = copy.deepcopy(self.rows)
        for frame in range(102, 105):
            rows.append(dict(rows[-1], frame=frame,
                             player=[rows[-1]['player'][0] + 3, rows[-1]['player'][1]],
                             target=[rows[-1]['target'][0] + 2, rows[-1]['target'][1] + 1]))
        samples = copy.deepcopy(self.samples)
        samples[-1][0]['pos'] = dict(zip(('x', 'y'), rows[-1]['player']))
        return rows, samples

    def test_contiguous_post_swap_movement_can_accumulate_more_than_exchange_tolerance(self):
        # Reduced from the real 627->628 exchange followed by seventeen
        # ordinary physics frames: total drift12.61, per-frame steps<4.68.
        rows, samples = self.moving_tail()
        result = self.check(rows, samples)
        self.assertTrue(result['passed'])
        motion = result['pairs'][0]['post_swap_motion']
        self.assertEqual(motion['maximum_player_frame_step'], 3)
        self.assertGreater(motion['final_distance_from_exchange'], self.required[0]['tolerance'])
        self.assertEqual(motion['final_distance_from_last_trace'], 0)

    def test_unobserved_drift_and_post_swap_teleport_cannot_pass(self):
        rows, samples = self.moving_tail()
        self.assertFalse(self.check(rows[:2], samples)['passed'])
        for unit in ('player', 'target'):
            changed = copy.deepcopy(rows)
            changed[-1][unit][0] += 20
            final = copy.deepcopy(samples)
            if unit == 'player':
                final[-1][0]['pos'] = dict(zip(('x', 'y'), changed[-1]['player']))
            with self.subTest(unit=unit):
                self.assertFalse(self.check(changed, final)['passed'])

    def test_post_swap_trace_gap_or_changed_input_identity_cannot_pass(self):
        for change in ({'frame': 106}, {'cast_frame': 99}, {'script_id': 3}, {'extent': 18},
                       {'wire': 20}, {'key': 32}, {'key_queued': False}):
            rows, samples = self.moving_tail()
            rows[-1].update(change)
            with self.subTest(change=change):
                self.assertFalse(self.check(rows, samples)['passed'])

    def test_continuous_return_to_start_does_not_count_as_escape(self):
        rows = copy.deepcopy(self.rows)
        for frame, x in enumerate(range(39, -1, -3), start=102):
            rows.append(dict(rows[-1], frame=frame, player=[x, 22]))
        samples = copy.deepcopy(self.samples)
        samples[-1][0]['pos'] = dict(zip(('x', 'y'), rows[-1]['player']))
        self.assertFalse(self.check(rows, samples)['passed'])

    def test_no_exchange_or_only_one_moving_object_cannot_pass(self):
        for player, target in (([0, 0], [39, 21]), ([42, 22], [39, 21]), ([5, 2], [0, 0])):
            rows = copy.deepcopy(self.rows)
            rows[1].update(player=player, target=target)
            with self.subTest(player=player, target=target):
                self.assertFalse(self.check(rows)['passed'])

    def test_missing_or_wrong_actual_key_identity_cannot_pass(self):
        for actual in ('', self.input.replace('key_queued=true', 'key_queued=false'),
                       self.input.replace('key=33', 'key=32'), self.input.replace('wire=0x13', 'wire=0x12'),
                       self.input.replace('extent=17', 'extent=18'), self.input.replace('script_id=2', 'script_id=3')):
            with self.subTest(actual=actual):
                self.assertFalse(self.check(actual=actual)['passed'])

    def test_missing_discontinuous_or_changed_identity_frames_cannot_pass(self):
        for change in ({'frame': 100}, {'frame': 102}, {'cast_frame': 99}, {'script_id': 3},
                       {'extent': 18}, {'wire': 20}, {'key': 32}, {'key_queued': False}):
            rows = copy.deepcopy(self.rows)
            rows[1].update(change)
            with self.subTest(change=change):
                self.assertFalse(self.check(rows)['passed'])
        self.assertFalse(self.check(self.rows[:1])['passed'])

    def test_invalid_native_positions_and_frame_words_cannot_pass(self):
        for change in ({'player': [math.nan, 0]}, {'target': [math.inf, 0]}, {'player': [0]},
                       {'frame': -1}, {'frame': 1 << 32}, {'frame': True}, {'key_queued': 1}):
            rows = copy.deepcopy(self.rows)
            rows[1].update(change)
            with self.subTest(change=change):
                result = self.check(rows)
                self.assertFalse(result['passed'])
                self.assertTrue(result['malformed_position_observation'])

    def test_wrong_seed_dead_target_or_missing_final_player_cannot_pass(self):
        for kind in ('outside-seed', 'dead', 'missing-target', 'missing-final', 'lost-final-position'):
            samples = copy.deepcopy(self.samples)
            if kind == 'outside-seed': samples[0][0]['pos'] = {'x': 50, 'y': 20}
            elif kind == 'dead': samples[0][1]['flags'].append('DEAD')
            elif kind == 'missing-target': samples[0].pop()
            elif kind == 'missing-final': samples[-1] = []
            else: samples[-1][0]['pos'] = {'x': 0, 'y': 0}
            with self.subTest(kind=kind):
                self.assertFalse(self.check(samples=samples)['passed'])

    def test_unsigned_frame_wraparound_keeps_the_one_frame_requirement(self):
        rows = copy.deepcopy(self.rows)
        rows[0]['frame'], rows[1]['frame'] = 0xffffffff, 0
        self.assertTrue(self.check(rows)['passed'])

    def test_invalid_suite_constraints_fail_before_launch(self):
        for change in ({'key': 32}, {'player_script_id': 2}, {'target_script_ids': [True]}, {'target_script_ids': [2, 2]}, {'target_script_ids': []},
                       {'minimum_distance': 8}, {'tolerance': -1}, {'player_start': [math.nan, 0]}):
            pair = dict(self.required[0], **change)
            with self.subTest(change=change), self.assertRaises(ValueError):
                runner.validate_suite_fields({'cases': [{'required_swap_pairs': [pair]}]})

    def test_actual_key_may_select_either_declared_original_target(self):
        samples = copy.deepcopy(self.samples)
        rows = [dict(row, script_id=3, wire=20, extent=18) for row in self.rows]
        actual = self.input.replace('wire=0x13', 'wire=0x14').replace('extent=17', 'extent=18').replace('script_id=2', 'script_id=3')
        result = self.check(rows, samples, actual)
        self.assertTrue(result['passed'])
        self.assertEqual(result['pairs'][0]['selected_target_script_id'], 3)
        rows[1]['script_id'] = 2
        self.assertFalse(self.check(rows, samples, actual)['passed'])


if __name__ == '__main__':
    unittest.main()
