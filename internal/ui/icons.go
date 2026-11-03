package ui

import "github.com/egoist/mygo/ui"

func icon(svg string) *ui.SVG {
	return ui.MustParseSVG([]byte(svg))
}

// The dashboard's icons are FlClash's own glyphs, redrawn as SVG. They are
// drawn on a 24-unit square with a 1.8 stroke, round caps and joins, and
// the corners of a shape are rounded, as FlClash's glyphs are.
const glyphStroke = `stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"`

var (
	iconDashboard = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="7" height="7" rx="1.5"/><rect x="14" y="3" width="7" height="7" rx="1.5"/><rect x="3" y="14" width="7" height="7" rx="1.5"/><rect x="14" y="14" width="7" height="7" rx="1.5"/></svg>`)
	iconProxies   = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><g fill="currentColor" fill-rule="evenodd"><path d="M8.048 0a8 8 0 1 0 .001 16.001A8 8 0 0 0 8.048 0M8 15A7 7 0 1 1 8 1a7 7 0 0 1 0 14"/><path d="M2.959 2.684c-.27 1.344.735 3.399 2.872 3.899s1.218.084.886-.583c-.334-.667-.035-1.167.482-1.333c.518-.167.387.293 1.155-.667c.194-2-1.472-.027-1.638-.86C8.366 1.223 5.95.5 5.249.75s-2.02.589-2.29 1.934m3.914 5.01c-.124.055-.702.416-.901.666s-.286.778 0 1c.286.223-.016 1.279.755 1.473s1.543.307 1.572.917c.027.61-.072 2.027-.443 2.167c.5.277 1.717-1.195 2.145-1.973c.429-.777.572-1.889.543-2.167s.171-.86.257-1.61s-.57-.8-.655-1.133c-.087-.334-.203-.701-3.273.66M13.125 4s-2.012.861-.644 2.14c1.366 1.277 2.062 2.49 1.995 2.768c-.062.276-1.38 1.706-.623 1.594c.761-.111 1.322-1.611 1.443-2.501c.119-.89-.924-2.639-.967-2.777c-.036-.141-.722-1.209-1.204-1.224"/></g></svg>`)
	iconProfiles  = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5"><path d="M2 5a2 2 0 0 1 2 -2h16a2 2 0 0 1 2 2v14a2 2 0 0 1 -2 2H4a2 2 0 0 1 -2 -2Z"/><path d="M3 7h18"/><path d="M9 11h6"/><path d="M9 14h6"/><path d="M9 17h6"/></g></svg>`)
	iconActivity  = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 12h-4l-3 8L9 4l-3 8H2"/></svg>`)
	iconLogs      = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32"><path fill="currentColor" d="M28 16v6H4V6h11V4H4a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h8v4H8v2h16v-2h-4v-4h8a2 2 0 0 0 2-2v-6ZM18 28h-4v-4h4Zm0-24h12v2H18z"/><path fill="currentColor" d="M18 8h12v2H18zm0 4h6v2h-6z"/></svg>`)
	iconTools     = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/></svg>`)
	iconPower     = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M18.36 6.64a9 9 0 1 1-12.73 0"/><line x1="12" y1="2" x2="12" y2="12"/></svg>`)
	iconRefresh   = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12a9 9 0 1 1-2.64-6.36"/><polyline points="21 3 21 9 15 9"/></svg>`)
	iconPlus      = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="currentColor" d="M11 13H5v-2h6V5h2v6h6v2h-6v6h-2z"/></svg>`)
	iconTrash     = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18"/><path d="M8 6V4h8v2"/><path d="M19 6l-1 14H6L5 6"/><line x1="10" y1="11" x2="10" y2="17"/><line x1="14" y1="11" x2="14" y2="17"/></svg>`)
	iconChevron   = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" ` + glyphStroke + `><path d="M9.5 5.5 L16.5 12 L9.5 18.5"/></svg>`)
	iconDown      = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><polyline points="19 12 12 19 5 12"/></svg>`)
	iconUp        = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="19" x2="12" y2="5"/><polyline points="5 12 12 5 19 12"/></svg>`)
	iconX         = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>`)
	iconClock     = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="9"/><polyline points="12 7 12 12 15.5 14"/></svg>`)
	iconSignal    = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4.9 19.1a13 13 0 0 1 0-14.2"/><path d="M7.8 16.2a8 8 0 0 1 0-8.4"/><path d="M10.6 13.4a3 3 0 0 1 0-2.8"/><circle cx="12" cy="12" r="1"/><path d="M13.4 10.6a3 3 0 0 1 0 2.8"/><path d="M16.2 7.8a8 8 0 0 1 0 8.4"/><path d="M19.1 4.9a13 13 0 0 1 0 14.2"/></svg>`)
	iconGlobe     = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="9"/><path d="M3 12h18"/><path d="M12 3a15 15 0 0 1 0 18a15 15 0 0 1 0-18z"/></svg>`)
	iconFolder    = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/></svg>`)
	iconCopy      = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" ` + glyphStroke + `><rect x="8.5" y="8.5" width="12" height="12" rx="2.5"/><path d="M15.5 5.5 A2 2 0 0 0 13.5 3.5 H5.5 A2 2 0 0 0 3.5 5.5 V13.5 A2 2 0 0 0 5.5 15.5"/></svg>`)
	iconPencil    = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 20h9"/><path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4z"/></svg>`)
	iconShield    = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>`)
	iconAlert     = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10.3 3.9 1.8 18a2 2 0 0 0 1.7 3h17a2 2 0 0 0 1.7-3L13.7 3.9a2 2 0 0 0-3.4 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>`)
	iconServer    = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="8" rx="2"/><rect x="2" y="13" width="20" height="8" rx="2"/><line x1="6" y1="7" x2="6.01" y2="7"/><line x1="6" y1="17" x2="6.01" y2="17"/></svg>`)
	iconScroll    = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M8 3h10a2 2 0 0 1 2 2v14a2 2 0 0 0 2 2H8a2 2 0 0 1-2-2V5a2 2 0 0 0-2-2z"/><path d="M10 7h6M10 11h6M10 15h4"/></svg>`)
	iconInfo      = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="9"/><line x1="12" y1="11" x2="12" y2="16"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>`)
	iconSun       = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4"/></svg>`)
	iconMoon      = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8z"/></svg>`)

	// ---- the dashboard's own glyphs, redrawn from FlClash ------------------

	// iconStatus is a rounded panel with a pulse line in it.
	iconStatus = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" ` + glyphStroke + `><rect x="3" y="4.2" width="18" height="15.6" rx="4.5"/><path d="M6.4 12.3 L8.8 12.3 L10.4 8.2 L13.4 16.4 L15 12.3 L17.6 12.3"/></svg>`)

	// iconSpeed is a dial with a needle.
	iconSpeed = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" ` + glyphStroke + `><path d="M6.759 19.057 Q5.259 19.057 4.396 17.83 A8.8 8.8 0 1 1 19.604 17.83 Q18.741 19.057 17.241 19.057 Z"/><circle cx="12" cy="13.4" r="1.4" fill="currentColor" stroke="none"/><path d="M12 13.4 L15.47 9.26"/></svg>`)

	// iconMemory is a chip with pins.
	iconMemory = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" ` + glyphStroke + `><rect x="3.5" y="7.2" width="17" height="9.6" rx="2.5"/><path d="M8 7.2 L8 3.8M12 7.2 L12 3.8M16 7.2 L16 3.8M8 16.8 L8 20.2M12 16.8 L12 20.2M16 16.8 L16 20.2"/></svg>`)

	// iconDevices is a desktop with a phone beside it.
	iconDevices = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" ` + glyphStroke + `><path d="M11.4 14.5 L4 14.5 L4 4 L18.8 4 L18.8 6.4"/><path d="M2.9 18.2 L11.4 18.2"/><rect x="15" y="10" width="5.8" height="10.5" rx="2"/></svg>`)

	// iconVerified is the service status: a hand holding a service tray.
	iconVerified = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="currentColor" d="M17.5 14.5v-1H20q.213 0 .356.144t.144.357t-.144.356T20 14.5zm0 4v-1H20q.213 0 .356.144t.144.357t-.144.356T20 18.5zm-3 1q-.825 0-1.412-.587T12.5 17.5h-2v-3h2q0-.825.588-1.412T14.5 12.5h2v7zm-7.5-3q-1.458 0-2.479-1.021T3.5 13t1.021-2.479T7 9.5h1.5q.817 0 1.409-.591q.591-.592.591-1.409t-.591-1.409Q9.317 5.5 8.5 5.5H5q-.213 0-.356-.144T4.5 4.999t.144-.356T5 4.5h3.5q1.258 0 2.129.871T11.5 7.5t-.871 2.129T8.5 10.5H7q-1.031 0-1.766.736q-.734.735-.734 1.768t.735 1.764T7 15.5h2.5v1z"/></svg>`)

	// iconLayers is a stack of three sheets.
	iconLayers = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" ` + glyphStroke + `><path d="M12 3 L20.8 7.6 L12 12.2 L3.2 7.6 Z"/><path d="M3.2 12 L12 16.6 L20.8 12"/><path d="M3.2 16.4 L12 21 L20.8 16.4"/></svg>`)

	// iconSplit is the outbound mode: traffic branching to two routes.
	iconSplit = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" ` + glyphStroke + `><path d="M4.5 4 L12 11.5 L12 20.5"/><path d="M19.5 4 L12 11.5 L12 20.5"/><path d="M4.5 9.5 L4.5 4 L10 4"/><path d="M14 4 L19.5 4 L19.5 9.5"/></svg>`)

	// iconShuffle is the system proxy: traffic taking another route.
	iconShuffle = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" ` + glyphStroke + `><path d="M3 7 L7 7 L15 17 L20.5 17"/><path d="M3 17 L5 17 Q7 17 8.25 15.44 L8.69 14.88"/><path d="M13.31 9.12 L13.75 8.56 Q15 7 17 7 L20.5 7"/><path d="M17.5 4 L20.5 7 L17.5 10"/><path d="M17.5 14 L20.5 17 L17.5 20"/></svg>`)

	// iconVPN is the TUN: a globe of network nodes.
	iconVPN = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="currentColor" d="M12 4a8 8 0 1 0 8 8a8.003 8.003 0 0 0-8-8m-6 8a6 6 0 0 1 .105-1.095L9.6 14.4v.8a1.605 1.605 0 0 0 1.6 1.6v1.14A6.004 6.004 0 0 1 6 12m9.2 3.2h-.8v-2.4a.8.8 0 0 0-.8-.8H8.8v-1.6h1.6a.8.8 0 0 0 .8-.8V8h1.6a1.6 1.6 0 0 0 1.59-1.5a5.985 5.985 0 0 1 2.137 9.426A1.57 1.57 0 0 0 15.2 15.2M13 1a1 1 0 1 1-1-1a1 1 0 0 1 1 1m5 1a1 1 0 1 1-1-1a1 1 0 0 1 1 1m4 4a1 1 0 1 1-1-1a1 1 0 0 1 1 1m2 6a1 1 0 1 1-1-1a1 1 0 0 1 1 1m-2 6a1 1 0 1 1-1-1a1 1 0 0 1 1 1m-4 4a1 1 0 1 1-1-1a1 1 0 0 1 1 1m-5 1a1 1 0 1 1-1-1a1 1 0 0 1 1 1m-5-1a1 1 0 1 1-1-1a1 1 0 0 1 1 1m-4-4a1 1 0 1 1-1-1a1 1 0 0 1 1 1m-2-6a1 1 0 1 1-1-1a1 1 0 0 1 1 1m2-6a1 1 0 1 1-1-1a1 1 0 0 1 1 1m4-4a1 1 0 1 1-1-1a1 1 0 0 1 1 1"/></svg>`)

	// iconBolt is the lightning bolt.
	iconBolt = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" ` + glyphStroke + `><path d="M14.5 2.8 L5 13.7 L11.2 13.7 L9.5 21.2 L19 10.3 L12.8 10.3 Z"/></svg>`)

	// iconSliders is the settings.
	iconSliders = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" ` + glyphStroke + `><path d="M3.2 5.5 L10.4 5.5M18.6 5.5 L20.8 5.5"/><circle cx="14.5" cy="5.5" r="2.5"/><path d="M3.2 12 L4.4 12M12.6 12 L20.8 12"/><circle cx="8.5" cy="12" r="2.5"/><path d="M3.2 18.5 L11.4 18.5M19.6 18.5 L20.8 18.5"/><circle cx="15.5" cy="18.5" r="2.5"/></svg>`)

	// iconSync is the update.
	iconSync = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="currentColor" fill-rule="evenodd" d="M19.5 7.03V3H21v5.75c0 .69-.56 1.25-1.25 1.25H14V8.5h4.63a7.5 7.5 0 0 0-6.63-4a7.48 7.48 0 0 0-6.93 4.63l-1.385-.575A8.97 8.97 0 0 1 12 3c3.06 0 5.86 1.56 7.5 4.03M5.37 15.5a7.53 7.53 0 0 0 6.63 4a7.48 7.48 0 0 0 6.93-4.63l1.385.575A8.97 8.97 0 0 1 12 21a9.03 9.03 0 0 1-7.5-4.025V21H3v-5.75c0-.69.56-1.25 1.25-1.25H10v1.5z" clip-rule="evenodd"/></svg>`)

	// iconDots is the menu of a profile.
	iconDots = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512"><path fill="currentColor" d="M113.7 304C86.2 304 64 282.6 64 256c0-26.5 22.2-48 49.7-48 27.6 0 49.8 21.5 49.8 48 0 26.6-22.2 48-49.8 48z"/><path fill="currentColor" d="M256 304c-27.5 0-49.8-21.4-49.8-48 0-26.5 22.3-48 49.8-48 27.5 0 49.7 21.5 49.7 48 0 26.6-22.2 48-49.7 48z"/><path fill="currentColor" d="M398.2 304c-27.5 0-49.8-21.4-49.8-48 0-26.5 22.2-48 49.8-48 27.5 0 49.8 21.5 49.8 48 0 26.6-22.2 48-49.8 48z"/></svg>`)

	// iconWifi is the LAN.
	iconWifi = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" ` + glyphStroke + `><path d="M2.5 8.5 A15 15 0 0 1 21.5 8.5"/><path d="M5.5 12 A10.5 10.5 0 0 1 18.5 12"/><path d="M8.5 15.5 A6 6 0 0 1 15.5 15.5"/><circle cx="12" cy="19" r="1.2" fill="currentColor" stroke="none"/></svg>`)

	// iconGithub is the GitHub mark, in the violet the project uses.
	iconGithub = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="#7c3aed" d="M12 2A10 10 0 0 0 2 12c0 4.42 2.87 8.17 6.84 9.5c.5.08.66-.23.66-.5v-1.69c-2.77.6-3.36-1.34-3.36-1.34c-.46-1.16-1.11-1.47-1.11-1.47c-.91-.62.07-.6.07-.6c1 .07 1.53 1.03 1.53 1.03c.87 1.52 2.34 1.07 2.91.83c.09-.65.35-1.09.63-1.34c-2.22-.25-4.55-1.11-4.55-4.92c0-1.11.38-2 1.03-2.71c-.1-.25-.45-1.29.1-2.64c0 0 .84-.27 2.75 1.02c.79-.22 1.65-.33 2.5-.33s1.71.11 2.5.33c1.91-1.29 2.75-1.02 2.75-1.02c.55 1.35.2 2.39.1 2.64c.65.71 1.03 1.6 1.03 2.71c0 3.82-2.34 4.66-4.57 4.91c.36.31.69.92.69 1.85V21c0 .27.16.59.67.5C19.14 20.16 22 16.42 22 12A10 10 0 0 0 12 2"/></svg>`)

	// iconCheckCircle marks the profile in use.
	iconCheckCircle = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="currentColor" d="m10.6 13.8l-2.15-2.15q-.275-.275-.7-.275t-.7.275t-.275.7t.275.7L9.9 15.9q.3.3.7.3t.7-.3l5.65-5.65q.275-.275.275-.7t-.275-.7t-.7-.275t-.7.275zM12 22q-2.075 0-3.9-.788t-3.175-2.137T2.788 15.9T2 12t.788-3.9t2.137-3.175T8.1 2.788T12 2t3.9.788t3.175 2.137T21.213 8.1T22 12t-.788 3.9t-2.137 3.175t-3.175 2.138T12 22"/></svg>`)

	// iconBoltFilled is the filled bolt of the test button.
	iconBoltFilled = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="currentColor" d="M13.82 3a.5.5 0 0 0-.847-.36l-1.778 1.71a35.6 35.6 0 0 0-6.63 8.715a.5.5 0 0 0 .435.746h4.31V21a.5.5 0 0 0 .837.37l.795-.725a35.5 35.5 0 0 0 7.001-8.78l.492-.87a.5.5 0 0 0-.435-.747h-4.18z"/></svg>`)

	// iconClear is the eraser of the clear-logs button.
	iconClear = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48"><defs><mask id="zgoclear"><g fill="none" stroke-linejoin="round" stroke-width="4"><path stroke="#fff" stroke-linecap="round" d="M20 5.914h8v8h15v8H5v-8h15z" clip-rule="evenodd"/><path fill="#fff" stroke="#fff" d="M8 40h32V22H8z"/><path stroke="#000" stroke-linecap="round" d="M16 39.898v-5.984m8 5.984v-6m8 6v-5.984"/><path stroke="#fff" stroke-linecap="round" d="M12 40h24"/></g></mask></defs><path fill="currentColor" d="M0 0h48v48H0z" mask="url(#zgoclear)"/></svg>`)

	// iconUnlink is the broken link of the close-all button.
	iconUnlink = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15.356 18.758a8.06 8.06 0 0 1-8.121-1.993a8.06 8.06 0 0 1-2.038-7.972m13.118 1.978l4.153-4.153m-9.238-.932l4.153-4.154M1.631 22.369l5.604-5.603M1.318 1.318l21.363 21.364M6.329 6.266l3.74-3.74l11.405 11.405l-3.74 3.74"/></svg>`)

	// iconArrowUp and iconArrowDown are the traffic directions.
	iconArrowUp   = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" ` + glyphStroke + `><path d="M12 5 L12 19"/><path d="M5.5 11.5 L12 5 L18.5 11.5"/></svg>`)
	iconArrowDown = icon(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" ` + glyphStroke + `><path d="M12 5 L12 19"/><path d="M5.5 12.5 L12 19 L18.5 12.5"/></svg>`)
)
