import pathlib, re
p = pathlib.Path(r"c:\Users\sharm\AI_AVENGERS-1\backend-go\cmd\server\main.go")
t = p.read_text(encoding="utf-8", errors="ignore")
orig = t
added_import = False
if "ai_avengers/backend/internal/vacuum" not in t:
    t = t.replace('"ai_avengers/backend/internal/validation"', '"ai_avengers/backend/internal/vacuum"\n\t"ai_avengers/backend/internal/validation"')
    added_import = True

# Insert service init after postgres/redis connects - find logger.Info("AI Avengers backend starting") region
insert_after = "postgres, err := db.Connect"
if "vacuumSvc" not in t:
    t = t.replace(
        "postgres, err := db.Connect",
        "vacuumSvc := vacuum.NewService(postgres.Pool, logger)\n\t_ = vacuumSvc.Brain().Reload(ctx)\n\tpostgres, err := db.Connect"
    )
else:
    # already patched but fix duplicate postgres declaration if double insert caused issue
    pass

# Wire handler after adminGroup creation: find adminGroup := v1.Group("/admin")
if "vacuumHandler" not in t:
    t = t.replace(
        'adminGroup := v1.Group("/admin")',
        'adminGroup := v1.Group("/admin")\n\tvacuumHandler := vacuum.NewHandler(postgres.Pool, vacuumSvc, logger)\n\tvacuumHandler.Register(adminGroup)'
    )

if t != orig:
    p.write_text(t, encoding="utf-8")
    print("patched main.go", "import added" if added_import else "")
else:
    print("no change")
