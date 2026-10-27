import { describe, test, expect, beforeEach, vi } from 'vitest';
import { api, buildTokenUrl, initToken, keyToByte, token, type ApplyResp } from './api';

beforeEach(() => {
  sessionStorage.clear();
});

describe('token', () => {
  test('URL token wins', () => {
    expect(initToken('?token=abc123', false)).toBe('abc123');
  });
  test('DEV without URL token falls back to dev', () => {
    expect(initToken('', true)).toBe('dev');
  });
  test('production without URL token is null', () => {
    expect(initToken('', false)).toBeNull();
  });
  test('token.get reads sessionStorage', () => {
    sessionStorage.setItem('token', 'xyz');
    expect(token.get()).toBe('xyz');
  });
});

describe('buildTokenUrl', () => {
  test('appends token from token.get()', () => {
    sessionStorage.setItem('token', 'abc');
    expect(buildTokenUrl('/api/profiles')).toBe('/api/profiles?token=abc');
  });
  test('appends with & when query already present', () => {
    sessionStorage.setItem('token', 'abc');
    expect(buildTokenUrl('/api/profiles?x=1')).toBe('/api/profiles?x=1&token=abc');
  });
  test('no token leaves path unchanged', () => {
    expect(buildTokenUrl('/api/profiles')).toBe('/api/profiles');
  });
});

describe('keyToByte', () => {
  test('uppercase letter', () => {
    expect(keyToByte('U')).toBe(0x55);
  });
  test('lowercase letter uppercases', () => {
    expect(keyToByte('u')).toBe(0x55);
  });
  test('digit', () => {
    expect(keyToByte('0')).toBe(0x30);
  });
  test('F-key rejected', () => {
    expect(keyToByte('F1')).toBeNull();
  });
  test('arrow rejected', () => {
    expect(keyToByte('ArrowLeft')).toBeNull();
  });
});

describe('api.applyBinding', () => {
  test('parses applied/skipped', async () => {
    sessionStorage.setItem('token', 'abc');
    const raw: ApplyResp = {
      applied: [{ profileId: 'f4', backupPath: '/x.bak', diffSummary: '1 byte changed' }],
      skipped: [{ profileId: 'f5', reason: 'binding conflict' }],
    };
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => raw,
    });

    const result = await api.applyBinding({
      control: 'Ignition',
      binding: { key: 'U' },
      profileIds: ['f4', 'f5'],
      force: false,
      profileOverrides: [],
    });
    expect(result.applied).toHaveLength(1);
    expect(result.applied[0].profileId).toBe('f4');
    expect(result.skipped).toHaveLength(1);
    expect(result.skipped[0].profileId).toBe('f5');
  });
});

describe('api.planCopy / api.applyCopy', () => {
  test('planCopy POSTs /api/plan/copy and returns a PlanResp', async () => {
    sessionStorage.setItem('token', 'abc');
    const raw = {
      plans: [{ profileId: 'dst', conflicts: [], warning: '', diffSummary: '1 byte changed', force: false }],
    };
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => raw,
    });

    const result = await api.planCopy({
      sourceProfileId: 'src',
      controls: ['Ignition'],
      profileIds: ['dst'],
      force: false,
      profileOverrides: [],
    });
    const url = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0][0] as string;
    expect(url).toContain('/api/plan/copy');
    expect(result.plans).toHaveLength(1);
    expect(result.plans[0].profileId).toBe('dst');
  });

  test('applyCopy POSTs /api/apply/copy and returns an ApplyResp', async () => {
    sessionStorage.setItem('token', 'abc');
    const raw = {
      applied: [{ profileId: 'dst', backupPath: '/x.bak', diffSummary: '1 byte changed' }],
      skipped: [],
    };
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => raw,
    });

    const result = await api.applyCopy({
      sourceProfileId: 'src',
      controls: ['Ignition'],
      profileIds: ['dst'],
      force: false,
      profileOverrides: [],
    });
    const url = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls[0][0] as string;
    expect(url).toContain('/api/apply/copy');
    expect(result.applied).toHaveLength(1);
  });
});
