import { describe, test, expect, vi } from 'vitest';
import { mount, unmount, tick } from 'svelte';
import ProfilePage from './ProfilePage.svelte';
import type { ControlDTO, ProfileDTO } from '../lib/api';

const profile: ProfileDTO = {
  id: 'f4',
  displayName: 'F4',
  isBaseline: false,
  hasControls: true,
  hasCalib: true,
};

function control(partial: Partial<ControlDTO>): ControlDTO {
  return {
    name: '',
    displayName: '',
    kind: 'unbound',
    key: '',
    button: 0,
    buttons: [],
    deviceName: '',
    axisName: '',
    ...partial,
  };
}

const controls: ControlDTO[] = [
  control({ name: 'Ignition', displayName: 'Ignition', kind: 'key', key: 'U' }),
  control({ name: 'TearOffVisor', displayName: 'Tear Off Visor', kind: 'button', button: 7, buttons: [7, 33] }),
];

async function settle() {
  for (let i = 0; i < 5; i++) {
    await Promise.resolve();
    await new Promise((r) => setTimeout(r, 0));
    await tick();
  }
}

function mountPage() {
  const target = document.createElement('div');
  document.body.appendChild(target);
  const onEditControl = vi.fn();
  const app = mount(ProfilePage, {
    target,
    props: { profile, profiles: [profile], controls, onEditControl },
  });
  return { target, app, onEditControl };
}

describe('ProfilePage', () => {
  test('default tab shows the controls table', () => {
    const { target, app } = mountPage();
    expect(target.querySelector('table')).toBeTruthy();
    unmount(app);
    target.remove();
  });

  test('header shows the control and bound counts', () => {
    const { target, app } = mountPage();
    const text = target.textContent ?? '';
    expect(text).toContain('2 controls');
    expect(text).toContain('2 bound');
    unmount(app);
    target.remove();
  });

  test('calibration tab shows CalibCopy', async () => {
    const { target, app } = mountPage();
    const calibTab = [...target.querySelectorAll('nav button')].find((b) =>
      b.textContent?.trim() === 'Calibration',
    );
    (calibTab as HTMLButtonElement).click();
    await settle();
    expect(target.querySelector('table')).toBeFalsy();
    expect(target.textContent).toContain('Copy calibration');
    unmount(app);
    target.remove();
  });

  test('edit button calls onEditControl', () => {
    const { target, app, onEditControl } = mountPage();
    const edit = [...target.querySelectorAll('button')].find((b) => b.textContent?.trim() === 'Edit');
    (edit as HTMLButtonElement).click();
    expect(onEditControl).toHaveBeenCalledWith('Ignition');
    unmount(app);
    target.remove();
  });

  test('copy selected opens the copy modal', async () => {
    const { target, app } = mountPage();
    const checkboxes = [...target.querySelectorAll('tbody input[type=checkbox]')] as HTMLInputElement[];
    checkboxes[0].checked = true;
    checkboxes[0].dispatchEvent(new Event('change', { bubbles: true }));
    await settle();

    const copyBtn = [...target.querySelectorAll('button')].find((b) =>
      b.textContent?.trim().includes('Copy selected'),
    );
    (copyBtn as HTMLButtonElement).click();
    await settle();
    expect(target.textContent).toContain('Copy bindings');
    unmount(app);
    target.remove();
  });

  test('copy selected with nothing selected shows a hint instead of silently doing nothing', async () => {
    const { target, app } = mountPage();
    const copyBtn = [...target.querySelectorAll('button')].find((b) =>
      b.textContent?.trim().includes('Copy selected'),
    ) as HTMLButtonElement;
    copyBtn.click();
    await settle();
    expect(target.textContent).toContain('Select at least one control');
    expect(target.textContent).not.toContain('Copy bindings');
    unmount(app);
    target.remove();
  });
});
