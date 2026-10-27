import { describe, test, expect, vi, beforeEach } from 'vitest';
import { mount, unmount, tick } from 'svelte';
import CopyModal from './CopyModal.svelte';
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
  for (let i = 0; i < 5; i++) {
    await Promise.resolve();
    await new Promise((r) => setTimeout(r, 0));
    await tick();
  }
}

function mountModal(props: Partial<Parameters<typeof CopyModal>[0]> = {}) {
  const target = document.createElement('div');
  document.body.appendChild(target);
  const onClose = vi.fn();
  const app = mount(CopyModal, {
    target,
    props: {
      profiles,
      sourceId: 'f4',
      controls: ['Ignition'],
      onClose,
      ...props,
    },
  });
  return { target, app, onClose };
}

describe('CopyModal', () => {
  beforeEach(() => {
    sessionStorage.setItem('token', 'abc');
  });

  test('source profile is not listed as a target', () => {
    const { target, app } = mountModal();
    const labels = [...target.querySelectorAll('label')].map((l) => l.textContent ?? '');
    const hasF4 = labels.some((l) => l.includes('F4'));
    expect(hasF4).toBe(false);
    const hasBaseline = labels.some((l) => l.includes('Baseline'));
    expect(hasBaseline).toBe(true);
    unmount(app);
    target.remove();
  });

  test('select all toggles targets', async () => {
    const { target, app } = mountModal();
    const selectAll = target.querySelector('.select-all input') as HTMLInputElement;
    const boxes = () =>
      [...target.querySelectorAll('label:not(.select-all) input')] as HTMLInputElement[];
    expect(boxes().every((b) => b.checked)).toBe(true);

    selectAll.click();
    await settle();
    expect(boxes().every((b) => b.checked)).toBe(false);

    selectAll.click();
    await settle();
    expect(boxes().every((b) => b.checked)).toBe(true);
    unmount(app);
    target.remove();
  });

  test('preview calls planCopy with targets excluding source', async () => {
    globalThis.fetch = vi.fn(async () =>
      jsonResponse({ plans: [{ profileId: 'Baseline', conflicts: [], warning: '', diffSummary: '', force: false }] }),
    ) as typeof fetch;

    const { target, app } = mountModal();
    const preview = [...target.querySelectorAll('button')].find((b) =>
      b.textContent?.trim() === 'Preview',
    );
    (preview as HTMLButtonElement).click();
    await settle();

    const calls = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls;
    const planCall = calls.find((c) => String(c[0]).includes('/api/plan/copy'));
    expect(planCall).toBeTruthy();
    const body = JSON.parse(planCall![1]!.body as string);
    expect(body.sourceProfileId).toBe('f4');
    expect(body.controls).toEqual(['Ignition']);
    expect(body.profileIds).toEqual(['Baseline', 'f5']);
    unmount(app);
    target.remove();
  });

  test('apply calls applyCopy', async () => {
    globalThis.fetch = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      if (String(input).includes('/api/apply/copy')) {
        return jsonResponse({ applied: [{ profileId: 'Baseline', backupPath: '', diffSummary: '' }], skipped: [] });
      }
      return jsonResponse({ plans: [] });
    }) as typeof fetch;

    const { target, app } = mountModal();
    const preview = [...target.querySelectorAll('button')].find((b) => b.textContent?.trim() === 'Preview');
    (preview as HTMLButtonElement).click();
    await settle();

    const apply = [...target.querySelectorAll('button')].find((b) => b.textContent?.trim() === 'Apply');
    (apply as HTMLButtonElement).click();
    await settle();

    const calls = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls;
    expect(calls.some((c) => String(c[0]).includes('/api/apply/copy'))).toBe(true);
    unmount(app);
    target.remove();
  });
});
