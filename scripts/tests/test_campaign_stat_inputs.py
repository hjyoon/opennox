import copy
import importlib.util
import json
from pathlib import Path
import unittest


spec = importlib.util.spec_from_file_location(
    'campaign_stat_inputs', Path(__file__).resolve().parents[1] / 'run-campaign-e2e.py')
runner = importlib.util.module_from_spec(spec)
spec.loader.exec_module(runner)


class PlayerStatInputTests(unittest.TestCase):
    def setUp(self):
        path = Path(runner.__file__).parent / 'e2e/campaign/wizard3-player-stat-audit.json'
        self.case = json.loads(path.read_text())['cases'][0]
        self.inputs = copy.deepcopy(self.case['expected_stat_inputs'])
        self.observed = dict(self.inputs, present=True, game_flags=268569091)
        self.player = {'player': {'health': 33, 'maximum_health': 32},
                       'mana': 49, 'maximum_mana': 98, 'stat_inputs': self.observed}

    def test_captured_inputs_and_full_resource_contract_pass(self):
        runner.validate_suite_fields({'cases': [self.case]})
        self.assertEqual(runner.player_value_failures(self.player, self.case), [])

    def test_suite_pins_inputs_from_independent_original_fixture(self):
        path = Path(runner.__file__).parents[1] / 'src/server/testdata/player_read_values_wizard3_original.json'
        fixture = json.loads(path.read_text())
        self.assertEqual(self.inputs, {key: fixture['inputs'][key] for key in self.inputs})
        self.assertEqual(self.inputs['level'], 3)
        self.assertEqual(self.inputs['class'], 1)

    def test_missing_snapshot_or_presence_is_not_evidence(self):
        for actual in (None, {}, dict(self.observed, present=False), dict(self.observed, present=1)):
            with self.subTest(actual=actual):
                failures = runner.player_value_failures(dict(self.player, stat_inputs=actual), self.case)
                self.assertEqual([row['check'] for row in failures], ['expected_stat_inputs'])

    def test_each_stat_word_and_identity_must_match_exactly(self):
        for key in runner.STAT_INPUT_WORDS | {'class', 'level'}:
            actual = copy.deepcopy(self.observed)
            if key in runner.STAT_INPUT_WORDS:
                for index in range(4):
                    changed = copy.deepcopy(actual)
                    changed[key][index] ^= 1
                    with self.subTest(key=key, index=index):
                        self.assertFalse(runner.expected_player_values(dict(self.player, stat_inputs=changed), self.case))
            else:
                actual[key] += 1
                with self.subTest(key=key):
                    self.assertFalse(runner.expected_player_values(dict(self.player, stat_inputs=actual), self.case))

    def test_boolean_or_untyped_observed_numbers_are_rejected(self):
        for key, value in [('class', True), ('level', 3.0), ('game_flags', True),
                           ('game_flags', -1), ('game_flags', 1 << 32),
                           ('maximum_bits32', [True, 1, 1, 1])]:
            with self.subTest(key=key, value=value):
                self.assertFalse(runner.expected_player_values(
                    dict(self.player, stat_inputs=dict(self.observed, **{key: value})), self.case))

    def test_invalid_expectations_are_rejected_before_launch(self):
        invalid = [None, {}, dict(self.inputs, typo=1), dict(self.inputs, class_=1),
                   dict(self.inputs, level=-1), dict(self.inputs, level=True),
                   dict(self.inputs, **{'class': 3}), dict(self.inputs, **{'class': True}),
                   dict(self.inputs, base_bits32=[1, 2, 3]),
                   dict(self.inputs, maximum_bits32=[1, 2, 3, 1 << 32]),
                   dict(self.inputs, warrior_bits32=[1, 2, 3, -1]),
                   dict(self.inputs, multipliers_bits32=[1, 2, 3, 1.0])]
        for value in invalid:
            with self.subTest(value=value), self.assertRaises(ValueError):
                runner.validate_suite_fields({'cases': [dict(self.case, expected_stat_inputs=value)]})

    def test_original_maximum_outcomes_are_still_required(self):
        for field in ('maximum_health', 'maximum_mana'):
            player = copy.deepcopy(self.player)
            target = player['player'] if field == 'maximum_health' else player
            target[field] += 1
            self.assertFalse(runner.expected_player_values(player, self.case))

    def test_cases_without_stat_requirement_keep_existing_behavior(self):
        self.assertTrue(runner.expected_player_values(
            {'player': {'health': 33}, 'mana': 49}, {'expected_health': 33, 'minimum_mana': 40}))


if __name__ == '__main__':
    unittest.main()
