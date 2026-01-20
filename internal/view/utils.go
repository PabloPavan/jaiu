package view

func studentsFilterChipClass(current, value string) string {
	base := "status-chip inline-flex items-center gap-2 rounded-lg border px-3 py-1 text-xs font-semibold transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500/40"
	active := "border-blue-400/80 bg-blue-500/15 text-blue-100 shadow-[0_0_0_1px_rgba(59,130,246,0.35)]"
	neutral := "border-slate-800 bg-slate-950/60 text-slate-300 hover:border-blue-400/50 hover:text-white"
	if current == value {
		return base + " " + active
	}
	return base + " " + neutral
}

func statusStyle(status string) string {
	switch status {
	case "inactive":
		return "inline-flex items-center rounded-full border border-slate-700/70 bg-slate-900/60 px-3 py-1 text-[11px] font-semibold text-slate-400"
	case "suspended":
		return "inline-flex items-center rounded-full border border-amber-400/40 bg-amber-400/10 px-3 py-1 text-[11px] font-semibold text-amber-200"
	default:
		return "inline-flex items-center rounded-full border border-emerald-400/40 bg-emerald-400/10 px-3 py-1 text-[11px] font-semibold text-emerald-200"
	}
}
