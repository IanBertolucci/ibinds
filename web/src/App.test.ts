import { describe, test, expect, vi } from 'vitest';
import { mount, tick, unmount } from 'svelte';
import App from './App.svelte';

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

const profiles = [
  { id: 'Baseline', displayName: 'Baseline', isBaseline: true, hasControls: true, hasCalib: false },
  { id: 'f4', displayName: 'F4', isBaseline: false, hasControls: true, hasCalib: false },
  { id: 'f5', displayName: 'F5', isBaseline: false, hasControls: true, hasCalib: false },
];

function controlsFor(id: string) {
  return [
    {
      name: 'BrakeBiasInc',
      displayName: 'Brake Bias Inc',
      kind: 'button',
      key: '',
      button: id === 'f4' ? 5 : 7,
      buttons: [id === 'f4' ? 5 : 7],
      deviceName: '',
      axisName: '',
    },
  ];
}

function mockApi() {
  globalThis.fetch = vi.fn(async (input: RequestInfo | URL) => {
    const url = String(input);
    if (url.startsWith('/api/profiles/')) {
      const id = decodeURIComponent(url.split('/')[3].split('?')[0]);
      return jsonResponse(controlsFor(id));
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

describe('App profile navigation', () => {
  test('sidebar lists all profiles and selecting one shows its controls', async () => {
    const { target, app } = await setupApp();

    const sidebarButtons = [...target.querySelectorAll('aside .sidebar button')];
    expect(sidebarButtons.map((b) => b.textContent?.trim())).toEqual([
      'Baseline (baseline)',
      'F4',
      'F5',
    ]);

    const f4 = sidebarButtons.find((b) => b.textContent?.includes('F4')) as HTMLButtonElement;
    f4.click();
    await settle();

    const binding = (target.textContent ?? '').replace(/\s+/g, ' ');
    expect(binding).toContain('Btn 5');

    unmount(app);
    target.remove();
  });

  test('New binding opens the wizard', async () => {
    const { target, app } = await setupApp();

    const newBinding = [...target.querySelectorAll('button')].find((b) =>
      b.textContent?.trim() === 'New binding',
    ) as HTMLButtonElement;
    newBinding.click();
    await settle();

    expect(target.textContent).toContain('Capture key');

    unmount(app);
    target.remove();
  });

  test('settings gear opens the profiles directory dialog', async () => {
    const { target, app } = await setupApp();

    const gear = target.querySelector('button[aria-label="Settings"]') as HTMLButtonElement;
    gear.click();
    await settle();

    expect(target.querySelector('[role="dialog"]')).toBeTruthy();
    expect(target.textContent).toContain('Profiles directory');
    expect(target.querySelector('input[aria-label="Profiles directory"]')).toBeTruthy();

    unmount(app);
    target.remove();
  });

  test('Edit on a control opens the wizard preset with that control', async () => {
    const { target, app } = await setupApp();

    const edit = [...target.querySelectorAll('button')].find((b) => b.textContent?.trim() === 'Edit');
    (edit as HTMLButtonElement).click();
    await settle();

    const select = target.querySelector('.modal select') as HTMLSelectElement;
    expect(select.value).toBe('BrakeBiasInc');

    unmount(app);
    target.remove();
  });
});

describe('App calibration tab', () => {
  test('calibration copy excludes the active profile from targets', async () => {
    const { target, app } = await setupApp();

    const f4 = [...target.querySelectorAll('aside .sidebar button')].find((b) =>
      b.textContent?.includes('F4'),
    ) as HTMLButtonElement;
    f4.click();
    await settle();

    const calibTab = [...target.querySelectorAll('.profile-page nav button')].find(
      (b) => b.textContent?.trim() === 'Calibration',
    ) as HTMLButtonElement;
    calibTab.click();
    await settle();

    const labels = [...target.querySelectorAll('.calib label')].map((l) => l.textContent ?? '');
    expect(labels.some((l) => l.includes('Baseline'))).toBe(true);
    expect(labels.some((l) => l.includes('F4'))).toBe(false);

    const targetChecks = [
      ...target.querySelectorAll('.target-row input[type=checkbox]'),
    ] as HTMLInputElement[];
    expect(targetChecks).toHaveLength(2);
    expect(targetChecks.every((b) => b.checked)).toBe(true);

    unmount(app);
    target.remove();
  });
});
