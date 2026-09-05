function AdminSettings() {
  // WHY this page is still a placeholder, unlike its 3 siblings built
  // in this same commit batch: GET/PATCH /admin/settings work against
  // an arbitrary JSONB `value` column per system_settings row (see
  // AdminHandler.GetSettings/UpdateSetting) - the 4 known setting keys
  // (china_wall, context, models, cost_budget) each have a COMPLETELY
  // DIFFERENT shape inside that JSONB blob (confirmed from the seed
  // data in AI_AVENGERS_SYSTEM_ARCHITECTURE.md section 5's INSERT
  // statement). A generic "edit arbitrary JSON" UI would be unsafe
  // (a typo could silently break China Wall enforcement in production
  // with zero client-side validation), while a bespoke form per known
  // key is real, scoped work deserving its own dedicated pass rather
  // than being rushed alongside 3 other pages in one commit.
  return (
    <div className="p-6">
      <h1 className="mb-2 text-xl font-semibold">Settings</h1>
      <p className="text-sm text-text-secondary">
        Per-key settings editor pending - see this file's header comment for why a generic JSON
        editor was deliberately not built here.
      </p>
    </div>
  )
}

export const Component = AdminSettings
