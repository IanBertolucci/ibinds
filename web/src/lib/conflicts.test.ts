import { describe, test, expect } from 'vitest';
import { highlight } from './conflicts';
import type { PlanResp } from './api';

describe('highlight', () => {
  test('groups conflicts by profileId', () => {
    const plan: PlanResp = {
      plans: [
        {
          profileId: 'f4',
          conflicts: [{ control: 'TractionControlArm', reason: "key 'U' already bound" }],
          warning: '',
          diffSummary: '',
          force: false,
        },
        { profileId: 'f5', conflicts: [], warning: '', diffSummary: '', force: false },
      ],
    };
    const h = highlight(plan);
    expect(h['f4']).toHaveLength(1);
    expect(h['f5']).toBeUndefined();
  });

  test('conflict-free plan yields empty object', () => {
    const plan: PlanResp = {
      plans: [
        { profileId: 'f4', conflicts: [], warning: '', diffSummary: '', force: false },
      ],
    };
    expect(highlight(plan)).toEqual({});
  });
});
