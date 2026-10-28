import { describe, test, expect, vi, beforeEach } from 'vitest';
import { mount, unmount, tick } from 'svelte';
import BindingModal from './BindingModal.svelte';
import type { ControlDTO, ProfileDTO } from '../lib/api';

const profiles: ProfileDTO[] = [
  { id: 'Baseline', displayName: 'Baseline', isBaseline: true, hasControls: true, hasCalib: true },
  { id: 'f4', displayName: 'F4', isBaseline: false, hasControls: true, hasCalib: true },
  { id: 'f5', displayName: 'F5', isBaseline: false, hasControls: true, hasCalib: true },
];

function control(name: string): ControlDTO {
  return { name, displayName: name, category: '', subcategory: '', kind: 'unbound', key: '', button: 0, buttons: [], deviceName: '', axisName: '' };
}

const controls: ControlDTO[] = [control('Ignition'), control('TractionControlArm')];

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

describe('BindingModal', () => {
  beforeEach(() => {
    sessionStorage.setItem('token', 'abc');
  });

  test('preset control is preselected in the wizard', async () => {
    globalThis.fetch = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes('/api/input/devices')) return jsonResponse([]);
      if (url.includes('/api/profiles/')) return jsonResponse(controls);
      if (url.includes('/api/profiles')) return jsonResponse(profiles);
      return jsonResponse({});
    }) as typeof fetch;

    const target = document.createElement('div');
    document.body.appendChild(target);
    const app = mount(BindingModal, {
      target,
      props: { profiles, controls, presetControl: 'Ignition', onClose: vi.fn() },
    });
    await settle();

    expect(target.querySelector('[role="dialog"]')).toBeTruthy();
    const select = target.querySelector('select') as HTMLSelectElement;
    expect(select.value).toBe('Ignition');

    unmount(app);
    target.remove();
  });

  test('preset default device is preselected in the wizard', async () => {
    const devices = [
      { name: 'Alpha Wheel', instanceGUID: '{AAA}', productGUID: 'p1', buttonCount: 16 },
      { name: 'Beta Wheel', instanceGUID: '{BBB}', productGUID: 'p2', buttonCount: 16 },
    ];
    globalThis.fetch = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes('/api/input/devices')) return jsonResponse(devices);
      if (url.includes('/api/profiles/')) return jsonResponse(controls);
      if (url.includes('/api/profiles')) return jsonResponse(profiles);
      return jsonResponse({});
    }) as typeof fetch;

    const target = document.createElement('div');
    document.body.appendChild(target);
    const app = mount(BindingModal, {
      target,
      props: {
        profiles,
        controls,
        presetControl: 'Ignition',
        defaultDeviceInstanceGUID: '{BBB}',
        onClose: vi.fn(),
      },
    });
    await settle();

    const deviceLabel = [...target.querySelectorAll('.manual-grid label.field')].find((l) =>
      l.textContent?.includes('Device'),
    );
    const select = deviceLabel?.querySelector('select') as HTMLSelectElement;
    expect(select.value).toBe('{BBB}');

    unmount(app);
    target.remove();
  });

  test('plan and apply use the preset control', async () => {
    globalThis.fetch = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      if (url.includes('/api/input/devices')) return jsonResponse([]);
      if (url.includes('/api/profiles/')) return jsonResponse(controls);
      if (url.includes('/api/profiles')) return jsonResponse(profiles);
      if (url.includes('/api/plan/binding')) {
        return jsonResponse({ plans: [{ profileId: 'f4', conflicts: [], warning: '', diffSummary: '', force: false }] });
      }
      if (url.includes('/api/apply/binding')) {
        return jsonResponse({ applied: [{ profileId: 'f4', backupPath: '', diffSummary: '' }], skipped: [] });
      }
      return jsonResponse({});
    }) as typeof fetch;

    const target = document.createElement('div');
    document.body.appendChild(target);
    const app = mount(BindingModal, {
      target,
      props: { profiles, controls, presetControl: 'Ignition', onClose: vi.fn() },
    });
    await settle();

    // Set a key binding via the dropdown, then Preview.
    const keyButton = [...target.querySelectorAll('button')].find((b) => b.textContent?.trim() === 'Use key');
    (keyButton as HTMLButtonElement).click();
    await settle();

    const preview = [...target.querySelectorAll('button')].find((b) => b.textContent?.trim() === 'Preview');
    (preview as HTMLButtonElement).click();
    await settle();

    const calls = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls;
    const planCall = calls.find((c) => String(c[0]).includes('/api/plan/binding'));
    expect(planCall).toBeTruthy();
    const body = JSON.parse(planCall![1]!.body as string);
    expect(body.control).toBe('Ignition');

    const apply = [...target.querySelectorAll('button')].find((b) => b.textContent?.trim() === 'Apply');
    (apply as HTMLButtonElement).click();
    await settle();

    const applyCall = (globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls.find((c) =>
      String(c[0]).includes('/api/apply/binding'),
    );
    expect(applyCall).toBeTruthy();

    unmount(app);
    target.remove();
  });
});
