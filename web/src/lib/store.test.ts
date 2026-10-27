import { describe, test, expect } from 'vitest';
import {
  buildBindingReq,
  currentBinding,
  loadControlsByProfile,
  selectAll,
  toggleSelection,
} from './store';
import type { ControlDTO, ProfileDTO } from './api';

function control(partial: Partial<ControlDTO>): ControlDTO {
  return {
    name: '',
    displayName: '',
    kind: 'unbound',
    key: '',
    button: 0,
    deviceName: '',
    axisName: '',
    ...partial,
  };
}

describe('selection', () => {
  test('select all and deselect one', () => {
    const sel = selectAll(['f4', 'f5']);
    expect(sel).toEqual({ f4: true, f5: true });
    expect(toggleSelection(sel, 'f4')).toEqual({ f4: false, f5: true });
  });
});

describe('buildBindingReq', () => {
  test('forceOverrides become profileOverrides', () => {
    const req = buildBindingReq('Ignition', { key: 'U' }, ['f4', 'f5'], false, {
      f4: true,
      f5: false,
    });
    expect(req.force).toBe(false);
    expect(req.profileOverrides).toEqual([{ profileId: 'f4', force: true }]);
  });
});

describe('loadControlsByProfile', () => {
  function profile(id: string, hasControls = true): ProfileDTO {
    return { id, displayName: id, isBaseline: false, hasControls, hasCalib: false };
  }

  test('loads controls for every profile, not only selected ones', async () => {
    const calls: string[] = [];
    const result = await loadControlsByProfile(
      [profile('f4'), profile('f5'), profile('f6')],
      async (id) => {
        calls.push(id);
        return [control({ name: 'Ignition', kind: 'button', button: 5 })];
      },
    );

    expect(calls.sort()).toEqual(['f4', 'f5', 'f6']);
    expect(result.f4[0].button).toBe(5);
    expect(result.f5[0].button).toBe(5);
    expect(result.f6[0].button).toBe(5);
  });

  test('is independent of hasControls and ignores fetch failures', async () => {
    const result = await loadControlsByProfile(
      [profile('has'), profile('missing', false), profile('broken')],
      async (id) => {
        if (id === 'broken') throw new Error('boom');
        return [control({ name: 'Ignition', kind: 'key', key: 'U' })];
      },
    );

    expect(Object.keys(result)).toEqual(['has', 'missing']);
  });
});

describe('currentBinding', () => {
  test('key', () => {
    expect(currentBinding(control({ kind: 'key', key: 'U' }))).toBe('Key U');
  });
  test('button with device', () => {
    expect(
      currentBinding(control({ kind: 'button', button: 128, deviceName: 'MOZA R9 Base' })),
    ).toBe('Btn 128 (MOZA R9 Base)');
  });
  test('unbound', () => {
    expect(currentBinding(control({ kind: 'unbound' }))).toBe('none');
  });
  test('axis', () => {
    expect(currentBinding(control({ kind: 'axis' }))).toBe('axis');
  });
});
