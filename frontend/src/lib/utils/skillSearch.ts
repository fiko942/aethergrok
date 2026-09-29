import type { SkillItem } from '../../app.d';

/**
 * Calculates a prioritized relevance score for a skill matching a search query.
 * Priority hierarchy:
 * 1. Exact Name match (1000)
 * 2. Name starts with query (800 - name.length)
 * 3. Name word / kebab-case segment starts with query (600 - name.length) (e.g. query "superpowers" matches segment in "using-superpowers")
 * 4. Name includes query anywhere (400 - name.length)
 * 5. ID / Scope matches (300)
 * 6. Tags or Actions match (250 for prefix, 200 for includes)
 * 7. Description word starts with query (100)
 * 8. Description includes query anywhere (50)
 */
export function getSkillMatchScore(skill: SkillItem, rawQuery: string): number {
  const cleanQuery = (rawQuery || '').toLowerCase().replace(/^\//, '').trim();
  if (!cleanQuery) return 1;

  const name = (skill.name || '').toLowerCase();
  const id = (skill.id || '').toLowerCase();
  const desc = (skill.description || '').toLowerCase();

  // 1. Exact match on name
  if (name === cleanQuery) {
    return 1000;
  }

  // 2. Name starts with query
  if (name.startsWith(cleanQuery)) {
    return 800 - name.length;
  }

  // 3. Name word / kebab-case / snake_case segment starts with query
  const segments = name.split(/[-_\s:/.]+/);
  for (const seg of segments) {
    if (seg.startsWith(cleanQuery)) {
      return 600 - name.length;
    }
  }

  // 4. Name contains query anywhere
  if (name.includes(cleanQuery)) {
    return 400 - name.length;
  }

  // 5. ID / Scope contains query
  if (id.includes(cleanQuery)) {
    return 300;
  }

  // 6. Tags or Actions match
  if (skill.tags && Array.isArray(skill.tags)) {
    for (const tag of skill.tags) {
      const t = tag.toLowerCase();
      if (t.startsWith(cleanQuery)) return 250;
      if (t.includes(cleanQuery)) return 200;
    }
  }
  if (skill.actions && Array.isArray(skill.actions)) {
    for (const act of skill.actions) {
      const a = act.toLowerCase();
      if (a.startsWith(cleanQuery)) return 250;
      if (a.includes(cleanQuery)) return 200;
    }
  }

  // 7. Description word starts with query
  const descWords = desc.split(/[\s,.-_()[\]]+/);
  for (const w of descWords) {
    if (w.startsWith(cleanQuery)) {
      return 100;
    }
  }

  // 8. Description contains query anywhere
  if (desc.includes(cleanQuery)) {
    return 50;
  }

  return 0;
}

/**
 * Filter and sort a list of skills by relevance according to query and optional category filter.
 */
export function rankSkills(
  skills: SkillItem[],
  query: string,
  category: string = 'All',
  limit?: number
): SkillItem[] {
  const cleanQuery = (query || '').toLowerCase().replace(/^\//, '').trim();
  const targetCat = category.trim().toLowerCase();

  const categoryFiltered = skills.filter((skill) => {
    return targetCat === 'all' || !skill.category || skill.category.toLowerCase() === targetCat;
  });

  if (!cleanQuery) {
    return typeof limit === 'number' ? categoryFiltered.slice(0, limit) : categoryFiltered;
  }

  const scored: { skill: SkillItem; score: number }[] = [];
  for (const skill of categoryFiltered) {
    const score = getSkillMatchScore(skill, cleanQuery);
    if (score > 0) {
      scored.push({ skill, score });
    }
  }

  scored.sort((a, b) => {
    if (b.score !== a.score) {
      return b.score - a.score;
    }
    if (a.skill.name.length !== b.skill.name.length) {
      return a.skill.name.length - b.skill.name.length;
    }
    return a.skill.name.localeCompare(b.skill.name);
  });

  const results = scored.map((item) => item.skill);
  return typeof limit === 'number' ? results.slice(0, limit) : results;
}
