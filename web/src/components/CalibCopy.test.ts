import { describe, test, expect, vi, beforeEach } from 'vitest';
import { mount, unmount, tick } from 'svelte';
import CalibCopy from './CalibCopy.svelte';
import type { ProfileDTO } from '../lib/api';

const profiles: ProfileDTO[] = [
  { id: 'Baseline', displayName: 'Baseline', isBaseline: true, hasControls: true, hasCalib: true },
  { id: 'f4', displayName: 'F4', isBaseline: false, hasControls: true, hasCalib: true },
  { id: 'f5', displayName: 'F5', isBaseline: false, hasControls: true, hasCalib: true },
];

function jsonResponse(v: unknown) {
  return Promise.resolve({ ok: true, status: 200, json: async () => v } as Response);
}

async function settle() {
  for (let i = 0; i < 6; i++) {
    await Promise.resolve();
    await new Promise((r) => setTimeout(r, 0));
    await tick();
  }
}

function mountCalib() {
  const target = document.createElement('div');
  document.body.appendChild(target);
  const onToggle = vi.fn();
  const app = mount(CalibCopy, {
    target,
    props: {
      profiles,
      selected: { Baseline: true, f5: true },
      onToggle,
      source: 'f4',
    },
  });
  return { target, app, onToggle };
}

describe('CalibCopy', () => {
  beforeEach(() => {
    sessionStorage.setItem('token', 'abc');
  });

  test('preview calls planCalib with targets excluding the source', async () => {
    globalThis.fetch = vi.fn(async () =>
      jsonResponse({
        plans: [
          { profileId: 'Baseline', conflicts: [], warning: '', diffSummary: '1 byte changed', force: false },
        ],
      }),
    ) as typeof fetch;

    const { target, app } = mountCalib();
    const preview = [...target.querySelectorAll('button')].find(
      (b) => b.textContent?.trim() === 'Preview',
    ) as HTMLButtonElement;
    preview.click();
    await settle();

    const calls = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls;
    const planCall = calls.find((c) => String(c[0]).includes('/api/plan/calib'));
    expect(planCall).toBeTruthy();
    const body = JSON.parse(planCall![1]!.body as string);
    expect(body.sourceProfileId).toBe('f4');
    expect(body.profileIds).toEqual(['Baseline', 'f5']);
    expect(target.textContent).toContain('1 byte changed');

    unmount(app);
    target.remove();
  });

  test('apply calls applyCalib after a preview', async () => {
    globalThis.fetch = vi.fn(async (input: RequestInfo | URL) => {
      if (String(input).includes('/api/apply/calib')) {
        return jsonResponse({
          applied: [{ profileId: 'Baseline', backupPath: '', diffSummary: '1 byte changed' }],
          skipped: [],
        });
      }
      return jsonResponse({ plans: [] });
    }) as typeof fetch;

    const { target, app } = mountCalib();
    const preview = [...target.querySelectorAll('button')].find(
      (b) => b.textContent?.trim() === 'Preview',
    ) as HTMLButtonElement;
    preview.click();
    await settle();

    const apply = [...target.querySelectorAll('button')].find(
      (b) => b.textContent?.trim() === 'Apply',
    ) as HTMLButtonElement;
    apply.click();
    await settle();

    const calls = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls;
    expect(calls.some((c) => String(c[0]).includes('/api/apply/calib'))).toBe(true);

    unmount(app);
    target.remove();
  });
});
