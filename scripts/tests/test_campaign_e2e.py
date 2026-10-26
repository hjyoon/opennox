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


if __name__ == '__main__':
    unittest.main()
