import importlib.util
from pathlib import Path
import unittest

spec = importlib.util.spec_from_file_location(
    'campaign_e2e', Path(__file__).resolve().parents[1] / 'run-campaign-e2e.py')
runner = importlib.util.module_from_spec(spec)
spec.loader.exec_module(runner)


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


if __name__ == '__main__':
    unittest.main()
