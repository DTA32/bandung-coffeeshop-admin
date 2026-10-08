package templates

// Class strings copied 1:1 from bandung-coffeeshop-fe-console
// (src/components/admin/AdminCafeForm.tsx and the admin.* routes) so the
// console keeps its look. Tailwind scans this file (@source '../../view').
const (
	// page wrapper: <main mainCls><div pageCls|pageNarrowCls>
	mainCls       = "flex-1 bg-cream px-6 py-8 md:px-16"
	pageCls       = "mx-auto flex max-w-4xl flex-col gap-6"
	pageNarrowCls = "mx-auto flex max-w-3xl flex-col gap-6"
	pageWideCls   = "mx-auto flex max-w-6xl flex-col gap-6"

	backLinkCls = "flex w-fit items-center gap-1 text-sm text-moss hover:underline"
	h1Cls       = "m-0 text-2xl font-bold text-forest"
	introCls    = "m-0 text-sm text-bark"

	// form sections
	cardCls      = "bg-white rounded-2xl p-5 flex flex-col gap-4"
	cardTightCls = "bg-white rounded-2xl p-5 flex flex-col gap-3"
	sectionTitle = "text-base font-bold text-forest m-0"
	labelCls     = "text-xs font-semibold text-bark"
	inputCls     = "rounded-lg border border-grove-light p-2 text-sm text-forest focus:border-grove focus:outline-none bg-white"
	searchCls    = "rounded-lg border border-grove-light p-3 text-sm text-forest focus:border-grove focus:outline-none"
	errorBoxCls  = "rounded-xl bg-red-50 p-4 text-sm text-red-800"
	okBoxCls     = "rounded-xl bg-forest-lighter p-4 text-sm text-moss"
	rowErrorCls  = "m-0 text-sm text-red-600"
	listRowCls   = "flex items-center justify-between rounded-2xl bg-white p-4 hover:shadow-sm"
	checkRowCls  = "flex w-fit items-center gap-2"
	checkTextCls = "text-sm text-forest"

	// buttons
	btnNewCls        = "flex items-center gap-1 rounded-lg bg-forest px-4 py-2 text-sm font-semibold text-cream"
	btnAddCls        = "flex w-fit items-center gap-1 rounded-lg bg-forest px-4 py-2 text-sm font-semibold text-cream disabled:opacity-40"
	btnSubmitCls     = "rounded-lg bg-forest px-6 py-2 text-sm font-semibold text-cream disabled:opacity-40"
	btnSaveSmCls     = "flex items-center gap-1 rounded-lg bg-forest px-5 py-1.5 text-sm font-semibold text-cream disabled:opacity-40"
	btnCancelCls     = "rounded-lg border border-grove-light px-4 py-2 text-sm font-semibold text-forest"
	btnOutlineCls    = "flex w-fit items-center gap-1 rounded-lg border border-grove-light px-3 py-2 text-sm text-forest"
	btnOutlineSmCls  = "flex items-center justify-center gap-1 rounded-lg border border-grove-light px-3 py-1.5 text-sm font-semibold text-forest hover:bg-forest-lighter disabled:opacity-40"
	btnDeleteCls     = "flex items-center gap-1 rounded-lg border border-red-300 px-4 py-2 text-sm font-semibold text-red-600 hover:bg-red-50 disabled:opacity-40"
	btnDeleteSmCls   = "flex items-center gap-1 rounded-lg border border-red-300 px-3 py-1.5 text-sm font-semibold text-red-600 hover:bg-red-50 disabled:opacity-40"
	btnPrimarySmCls  = "flex items-center justify-center gap-1 rounded-lg bg-forest px-3 py-1.5 text-sm font-semibold text-cream disabled:opacity-40"
	btnIconCls       = "rounded-md border border-grove-light p-1 text-bark hover:text-forest"
	btnIconDeleteCls = "rounded-md border border-grove-light p-1 text-red-500 hover:bg-red-50"
	inlineLinkCls    = "text-moss hover:underline"

	// status pill (list rows)
	pillCls = "rounded-md px-2.5 py-1 text-xs font-semibold"
)

// statusColor mirrors the console's statusColor map.
func statusColor(status string) string {
	switch status {
	case "active", "approved", "live":
		return "bg-forest-lighter text-moss"
	case "closed", "rejected", "expired", "running":
		return "bg-amber-50 text-amber-700"
	case "deleted", "failed":
		return "bg-red-50 text-red-700"
	default:
		return "bg-grove-light text-forest"
	}
}

func sourceColor(source string) string {
	switch source {
	case "gmaps":
		return "bg-sky-50 text-sky-700"
	case "instagram":
		return "bg-pink-50 text-pink-700"
	case "linktree":
		return "bg-lime-50 text-lime-700"
	default:
		return "bg-grove-light text-forest"
	}
}

func sourceLabel(source string) string {
	switch source {
	case "gmaps":
		return "Google Maps"
	case "instagram":
		return "Instagram"
	case "linktree":
		return "Linktree"
	default:
		return source
	}
}
