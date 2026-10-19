import { describe, test, expect } from 'vitest';
import {
  buildBindingReq,
  currentBinding,
  selectAll,
  toggleSelection,
} from './store';
import type { ControlDTO } from './api';

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
