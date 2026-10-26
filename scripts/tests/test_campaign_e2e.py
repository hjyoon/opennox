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


class PlayerOutcomeTests(unittest.TestCase):
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
