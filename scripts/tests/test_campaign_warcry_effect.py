import copy
import importlib.util
from pathlib import Path
import unittest

spec = importlib.util.spec_from_file_location(
    'campaign_warcry', Path(__file__).resolve().parents[1] / 'run-campaign-e2e.py')
runner = importlib.util.module_from_spec(spec)
spec.loader.exec_module(runner)


class OriginalWarcryEffectTests(unittest.TestCase):
    def setUp(self):
        self.required = dict(script_id=1000347545, power=5, duration_frames=120,
                             before_checkpoint='before', active_checkpoint='active',
                             expired_checkpoint='expired')
        self.samples = []
        for name, frame, active, remaining in [('before', 162, False, 0),
                                                ('active', 206, True, 94),
                                                ('expired', 400, False, 0)]:
            obj = dict(ind=9816, script_id=1000347545, enabled=True,
                       flags=['ACTIVE', 'ENABLED'], **{'class': ['MONSTER', 'COMPLEX']},
                       health=dict(current=40, maximum=40))
            combat = dict(ind=9816, script_id=1000347545, buffs=(1 << 29) if active else 0,
                          anti_magic=dict(active=active, frames=remaining, power=5 if active else 0),
                          monster=dict(status_flags=40, can_cast_spells=True))
            self.samples.append(dict(checkpoint=name, objects=[obj],
                                     combat=dict(frame=frame, objects=[combat])))

    def check(self, samples=None):
        return runner.required_warcry_effect(self.samples if samples is None else samples,
                                            self.required)

    def test_living_original_native_caster_cycle_passes_without_mutating_samples(self):
        before = copy.deepcopy(self.samples)
        result = self.check()
        self.assertTrue(result['passed'])
        self.assertEqual(result['phases'][1]['anti_magic']['frames'], 94)
        self.assertEqual(self.samples, before)

    def test_missing_repeated_or_out_of_order_phases_fail(self):
        for samples in ([], self.samples[1:], self.samples + self.samples[:1],
                        self.samples[::-1]):
            with self.subTest(samples=samples):
                self.assertFalse(self.check(samples)['passed'])

    def test_missing_duplicate_or_reused_native_identity_fails(self):
        for route in ('objects', 'combat'):
            for change in ('missing', 'duplicate', 'reused'):
                samples = copy.deepcopy(self.samples)
                objects = samples[1]['objects'] if route == 'objects' else samples[1]['combat']['objects']
                if change == 'missing':
                    objects.clear()
                elif change == 'duplicate':
                    objects.append(copy.deepcopy(objects[0]))
                else:
                    objects[0]['ind'] += 1
                with self.subTest(route=route, change=change):
                    self.assertFalse(self.check(samples)['passed'])

    def test_no_missing_health_dead_disabled_or_non_monster_caster_credit(self):
        for phase in range(3):
            for change in ('missing-health', 'zero-health', 'DEAD', 'DESTROYED', 'disabled', 'player'):
                samples = copy.deepcopy(self.samples)
                obj = samples[phase]['objects'][0]
                if change == 'missing-health': del obj['health']
                elif change == 'zero-health': obj['health']['current'] = 0
                elif change in ('DEAD', 'DESTROYED'): obj['flags'].append(change)
                elif change == 'disabled': obj['enabled'] = False
                else: obj['class'] = ['PLAYER']
                with self.subTest(phase=phase, change=change):
                    self.assertFalse(self.check(samples)['passed'])

    def test_can_cast_flag_and_native_buff_bit_must_agree_with_observer(self):
        for change in ('missing-monster', 'caster-bool', 'caster-bit', 'buff-bit', 'buff-bool'):
            samples = copy.deepcopy(self.samples)
            state = samples[1]['combat']['objects'][0]
            if change == 'missing-monster': del state['monster']
            elif change == 'caster-bool': state['monster']['can_cast_spells'] = False
            elif change == 'caster-bit': state['monster']['status_flags'] = 8
            elif change == 'buff-bit': state['buffs'] = 0
            else: state['anti_magic']['active'] = False
            with self.subTest(change=change):
                self.assertFalse(self.check(samples)['passed'])

    def test_wrong_power_zero_or_excessive_duration_and_early_expiry_fail(self):
        for change in ('power', 'zero-duration', 'long-duration', 'early-expiry', 'already-active', 'still-active'):
            samples = copy.deepcopy(self.samples)
            active = samples[1]['combat']['objects'][0]['anti_magic']
            if change == 'power': active['power'] = 4
            elif change == 'zero-duration': active['frames'] = 0
            elif change == 'long-duration': active['frames'] = 121
            elif change == 'early-expiry': samples[2]['combat']['frame'] = 299
            else:
                state = samples[0 if change == 'already-active' else 2]['combat']['objects'][0]
                state['anti_magic'].update(active=True, frames=90, power=5)
                state['buffs'] = 1 << 29
            with self.subTest(change=change):
                self.assertFalse(self.check(samples)['passed'])

    def test_invalid_native_words_and_structures_fail_closed(self):
        for change in ('frame', 'bool-frame', 'duration', 'bool-duration', 'health', 'identity', 'buffs', 'objects'):
            samples = copy.deepcopy(self.samples)
            state = samples[1]['combat']['objects'][0]
            if change == 'frame': samples[1]['combat']['frame'] = -1
            elif change == 'bool-frame': samples[1]['combat']['frame'] = True
            elif change == 'duration': state['anti_magic']['frames'] = 94.0
            elif change == 'bool-duration': state['anti_magic']['frames'] = True
            elif change == 'health': samples[1]['objects'][0]['health']['current'] = True
            elif change == 'identity': samples[1]['objects'][0]['ind'] = 9816.0
            elif change == 'buffs': state['buffs'] = 1 << 32
            else: samples[1]['combat']['objects'] = [None]
            with self.subTest(change=change):
                self.assertFalse(self.check(samples)['passed'])

    def test_uint32_frame_wrap_preserves_duration_and_rejects_reversed_time(self):
        samples = copy.deepcopy(self.samples)
        for sample, frame in zip(samples, (0xfffffff0, 0x20, 0xb0)):
            sample['combat']['frame'] = frame
        self.assertTrue(self.check(samples)['passed'])
        samples[2]['combat']['frame'] = 0x1f
        self.assertFalse(self.check(samples)['passed'])

    def test_inactive_raw_counter_and_power_are_not_assumed_cleared(self):
        samples = copy.deepcopy(self.samples)
        for i in (0, 2):
            samples[i]['combat']['objects'][0]['anti_magic'].update(frames=17, power=5)
        self.assertTrue(self.check(samples)['passed'])

    def test_requirement_typos_invalid_words_and_ambiguous_phase_contracts_fail_before_launch(self):
        case = dict(expected_maps=['war03c'], required_checkpoints=['before', 'active', 'expired'],
                    required_warcry_effect=self.required)
        runner.validate_suite_fields({'cases': [case]})
        for change in ({'script_id': True}, {'script_id': 0}, {'power': 0}, {'duration_frames': 65536},
                       {'duration_frames': True}, {'before_checkpoint': 'active'},
                       {'active_checkpoint': 'unrequired'}, {'typo': 1}):
            with self.subTest(change=change), self.assertRaises(ValueError):
                runner.validate_suite_fields({'cases': [dict(case, required_warcry_effect=dict(self.required, **change))]})
        for maps in (None, [], ['war03b', 'war03c']):
            with self.subTest(maps=maps), self.assertRaises(ValueError):
                runner.validate_suite_fields({'cases': [dict(case, expected_maps=maps)]})


if __name__ == '__main__':
    unittest.main()
