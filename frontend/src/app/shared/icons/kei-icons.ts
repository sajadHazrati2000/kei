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

  'chevron-left':  icon(`<path d="M15 6l-6 6l6 6"/>`),
  'chevron-right': icon(`<path d="M9 6l6 6l-6 6"/>`),
  'pencil': icon(
    `<path d="M4 20h4l10.5 -10.5a2.828 2.828 0 1 0 -4 -4l-10.5 10.5v4"/>` +
    `<path d="M13.5 6.5l4 4"/>`
  ),
  'trash': icon(
    `<path d="M4 7l16 0"/><path d="M10 11l0 6"/><path d="M14 11l0 6"/>` +
    `<path d="M5 7l1 12a2 2 0 0 0 2 2h8a2 2 0 0 0 2 -2l1 -12"/>` +
    `<path d="M9 7v-3a1 1 0 0 1 1 -1h4a1 1 0 0 1 1 1v3"/>`
  ),
  'circle-check': icon(
    `<path d="M3 12a9 9 0 1 0 18 0a9 9 0 1 0 -18 0"/>` +
    `<path d="M9 12l2 2l4 -4"/>`
  ),
  'circle-x': icon(
    `<path d="M3 12a9 9 0 1 0 18 0a9 9 0 1 0 -18 0"/>` +
    `<path d="M10 10l4 4m0 -4l-4 4"/>`
  ),
  'focus-2': icon(
    `<path d="M11.5 12a.5 .5 0 1 0 1 0a.5 .5 0 1 0 -1 0" fill="currentColor"/>` +
    `<path d="M5 12a7 7 0 1 0 14 0a7 7 0 1 0 -14 0"/>` +
    `<path d="M12 3l0 2"/><path d="M3 12l2 0"/>` +
    `<path d="M12 19l0 2"/><path d="M19 12l2 0"/>`
  ),
  'x': icon(`<path d="M18 6l-12 12"/><path d="M6 6l12 12"/>`),

  'check': icon(
    `<path d="M5 12l5 5l10 -10"/>`
  ),

  'loader': icon(
    `<path d="M12 6l0 -3"/><path d="M16.25 7.75l2.15 -2.15"/>` +
    `<path d="M18 12l3 0"/><path d="M16.25 16.25l2.15 2.15"/>` +
    `<path d="M12 18l0 3"/><path d="M7.75 16.25l-2.15 2.15"/>` +
    `<path d="M6 12l-3 0"/><path d="M7.75 7.75l-2.15 -2.15"/>`
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
