// ─── Kei Icon Registry ────────────────────────────────────────────────────────
// SVG content sourced from @tabler/icons v3 (MIT licence).
// Icons are outline variants, 24×24 viewBox, stroke="currentColor".
// Only the paths used in this application are bundled here.

const SVG_OPEN =
  `<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" ` +
  `fill="none" stroke="currentColor" stroke-width="2" ` +
  `stroke-linecap="round" stroke-linejoin="round">`;
const SVG_CLOSE = `</svg>`;

function icon(paths: string): string {
  return `${SVG_OPEN}<path stroke="none" d="M0 0h24v24H0z" fill="none"/>${paths}${SVG_CLOSE}`;
}

export const KEI_ICONS: Record<string, string> = {
  'layout-dashboard': icon(
    `<path d="M5 4h4a1 1 0 0 1 1 1v6a1 1 0 0 1 -1 1h-4a1 1 0 0 1 -1 -1v-6a1 1 0 0 1 1 -1"/>` +
    `<path d="M5 16h4a1 1 0 0 1 1 1v2a1 1 0 0 1 -1 1h-4a1 1 0 0 1 -1 -1v-2a1 1 0 0 1 1 -1"/>` +
    `<path d="M15 12h4a1 1 0 0 1 1 1v6a1 1 0 0 1 -1 1h-4a1 1 0 0 1 -1 -1v-6a1 1 0 0 1 1 -1"/>` +
    `<path d="M15 4h4a1 1 0 0 1 1 1v2a1 1 0 0 1 -1 1h-4a1 1 0 0 1 -1 -1v-2a1 1 0 0 1 1 -1"/>`
  ),

  'calendar': icon(
    `<path d="M4 7a2 2 0 0 1 2 -2h12a2 2 0 0 1 2 2v12a2 2 0 0 1 -2 2h-12a2 2 0 0 1 -2 -2v-12"/>` +
    `<path d="M16 3v4"/><path d="M8 3v4"/><path d="M4 11h16"/>` +
    `<path d="M11 15h1"/><path d="M12 15v3"/>`
  ),

  'users': icon(
    `<path d="M5 7a4 4 0 1 0 8 0a4 4 0 1 0 -8 0"/>` +
    `<path d="M3 21v-2a4 4 0 0 1 4 -4h4a4 4 0 0 1 4 4v2"/>` +
    `<path d="M16 3.13a4 4 0 0 1 0 7.75"/>` +
    `<path d="M21 21v-2a4 4 0 0 0 -3 -3.85"/>`
  ),

  'settings': icon(
    `<path d="M10.325 4.317c.426 -1.756 2.924 -1.756 3.35 0a1.724 1.724 0 0 0 2.573 1.066` +
    `c1.543 -.94 3.31 .826 2.37 2.37a1.724 1.724 0 0 0 1.065 2.572c1.756 .426 1.756 2.924 0 3.35` +
    `a1.724 1.724 0 0 0 -1.066 2.573c.94 1.543 -.826 3.31 -2.37 2.37a1.724 1.724 0 0 0 -2.572 1.065` +
    `c-.426 1.756 -2.924 1.756 -3.35 0a1.724 1.724 0 0 0 -2.573 -1.066c-1.543 .94 -3.31 -.826 -2.37 -2.37` +
    `a1.724 1.724 0 0 0 -1.065 -2.572c-1.756 -.426 -1.756 -2.924 0 -3.35a1.724 1.724 0 0 0 1.066 -2.573` +
    `c-.94 -1.543 .826 -3.31 2.37 -2.37c1 .608 2.296 .07 2.572 -1.065"/>` +
    `<path d="M9 12a3 3 0 1 0 6 0a3 3 0 0 0 -6 0"/>`
  ),

  'user': icon(
    `<path d="M8 7a4 4 0 1 0 8 0a4 4 0 0 0 -8 0"/>` +
    `<path d="M6 21v-2a4 4 0 0 1 4 -4h4a4 4 0 0 1 4 4v2"/>`
  ),

  'language': icon(
    `<path d="M9 6.371c0 4.418 -2.239 6.629 -5 6.629"/>` +
    `<path d="M4 6.371h7"/>` +
    `<path d="M5 9c0 2.144 2.252 3.908 6 4"/>` +
    `<path d="M12 20l4 -9l4 9"/>` +
    `<path d="M19.1 18h-6.2"/>` +
    `<path d="M6.694 3l.793 .582"/>`
  ),
};
