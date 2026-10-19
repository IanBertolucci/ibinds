import type { ConflictDTO, PlanResp } from './api';

// highlight groups conflicts by profileId. Profiles with no conflicts are
// omitted (an empty object means no conflicts anywhere).
export function highlight(plan: PlanResp): Record<string, ConflictDTO[]> {
  const out: Record<string, ConflictDTO[]> = {};
  for (const p of plan.plans) {
    if (p.conflicts.length > 0) {
      out[p.profileId] = p.conflicts;
    }
  }
  return out;
}
