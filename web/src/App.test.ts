import { describe, test, expect, vi } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import App from './App.svelte';

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

const profiles = [
  { id: 'Baseline', displayName: 'Baseline', isBaseline: true, hasControls: true, hasCalib: false },
  { id: 'f4', displayName: 'F4', isBaseline: false, hasControls: true, hasCalib: false },
  { id: 'f5', displayName: 'F5', isBaseline: false, hasControls: true, hasCalib: false },
];

function mockApi() {
  globalThis.fetch = vi.fn(async (input: RequestInfo | URL) => {
    const url = String(input);
    if (url.startsWith('/api/profiles/')) {
      const id = decodeURIComponent(url.split('/')[3].split('?')[0]);
      return jsonResponse([
        {
          name: 'BrakeBiasInc',
          displayName: 'Brake Bias Inc',
          kind: 'button',
          key: '',
          button: id === 'f4' ? 5 : 7,
          deviceName: '',
          axisName: '',
        },
      ]);
    }
    if (url.startsWith('/api/profiles')) return jsonResponse(profiles);
    if (url.startsWith('/api/input/devices')) return jsonResponse([]);
    return jsonResponse({});
  }) as typeof fetch;
}

async function setupApp() {
  mockApi();
  const target = document.createElement('div');
  document.body.appendChild(target);
  const app = mount(App, { target });
  await settle();
  return { target, app };
}

describe('App control selection', () => {
  test('shows the current binding for every profile, selected or not', async () => {
    const { target, app } = await setupApp();

    const select = target.querySelector('select') as HTMLSelectElement;
    select.value = 'BrakeBiasInc';
    select.dispatchEvent(new Event('change', { bubbles: true }));
    await settle();

    const current = [...target.querySelectorAll('.profile-list .current')].map((n) => n.textContent?.trim());
    expect(current).toEqual(['Btn 7', 'Btn 5', 'Btn 7']);

    unmount(app);
    target.remove();
  });
});

describe('App calibration targets', () => {
  test('toggle all clears and restores every target', async () => {
    const { target, app } = await setupApp();

    const calibTab = [...target.querySelectorAll('nav button')].find(
      (b) => b.textContent?.trim() === 'Calibration',
    ) as HTMLButtonElement;
    calibTab.click();
    await settle();

    const targetBoxes = () =>
      [...target.querySelectorAll('.calib label:not(.select-all) input')] as HTMLInputElement[];
    const selectAll = target.querySelector('.calib label.select-all input') as HTMLInputElement;
    const targetCount = () =>
      [...target.querySelectorAll('.calib p')].find((p) => p.textContent?.startsWith('Targets'))?.textContent;

    expect(targetBoxes().map((b) => b.checked)).toEqual([true, true, true]);
    expect(targetCount()).toBe('Targets (3):');

    selectAll.click();
    await settle();
    expect(targetBoxes().map((b) => b.checked)).toEqual([false, false, false]);
    expect(targetCount()).toBe('Targets (0):');

    selectAll.click();
    await settle();
    expect(targetBoxes().map((b) => b.checked)).toEqual([true, true, true]);
    expect(targetCount()).toBe('Targets (3):');

    unmount(app);
    target.remove();
  });
});

