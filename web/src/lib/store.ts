import type { BindingReq, ControlDTO, ProfileDTO } from './api';

export type Selection = Record<string, boolean>;

// selectAll marks every profile id as selected.
export function selectAll(profileIds: string[]): Selection {
  const s: Selection = {};
  for (const id of profileIds) s[id] = true;
  return s;
}

// toggleSelection flips one profile's selection state.
export function toggleSelection(sel: Selection, id: string): Selection {
  return { ...sel, [id]: !sel[id] };
}

// buildProfileOverrides turns per-profile force flags into the
// profileOverrides list the API expects.
export function buildProfileOverrides(
  forceOverrides: Record<string, boolean>,
): { profileId: string; force: boolean }[] {
  return Object.entries(forceOverrides)
    .filter(([, force]) => force)
    .map(([profileId, force]) => ({ profileId, force }));
}

// buildBindingReq assembles the request body from the wizard state.
export function buildBindingReq(
  control: string,
  binding: BindingReq['binding'],
  profileIds: string[],
  force: boolean,
  forceOverrides: Record<string, boolean>,
): BindingReq {
  return {
    control,
    binding,
    profileIds,
    force,
    profileOverrides: buildProfileOverrides(forceOverrides),
  };
}

// loadControlsByProfile fetches the control list for every profile. It
// deliberately ignores both the selection state and hasControls: the UI must
// show each profile's current binding regardless of whether the profile is
// selected, and a failed/absent controls.cfg is simply skipped.
export async function loadControlsByProfile(
  profiles: ProfileDTO[],
  fetchControls: (profileId: string) => Promise<ControlDTO[]>,
): Promise<Record<string, ControlDTO[]>> {
  const results = await Promise.all(
    profiles.map(async (p): Promise<[string, ControlDTO[]] | null> => {
      try {
        return [p.id, await fetchControls(p.id)];
      } catch {
        return null;
      }
    }),
  );
  const byProfile: Record<string, ControlDTO[]> = {};
  for (const entry of results) {
    if (entry) byProfile[entry[0]] = entry[1];
  }
  return byProfile;
}

// currentBinding renders a control's current binding as human text.
export function currentBinding(control: ControlDTO): string {
  switch (control.kind) {
    case 'key':
      return `Key ${control.key}`;
    case 'button':
      return control.deviceName
        ? `Btn ${control.button} (${control.deviceName})`
        : `Btn ${control.button}`;
    case 'axis':
      return 'axis';
    default:
      return 'none';
  }
}
