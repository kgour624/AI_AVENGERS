import pathlib
p = pathlib.Path(r"C:\Users\sharm\AI_AVENGERS-1\backend-go\internal\admin\audit_log.go")
txt = p.read_text(encoding="utf-8", errors="ignore")
# rename helpers to avoid duplicate with any existing itoa in admin pkg
txt = txt.replace("func atoiBounded", "func auditAtoiBounded").replace("func itoa", "func auditItoa")
txt = txt.replace("atoiBounded(", "auditAtoiBounded(").replace("itoa(", "auditItoa(")
# also fix parseErr name to auditParseErr
txt = txt.replace("type parseErr", "type auditParseErr").replace("&parseErr{", "&auditParseErr{")
p.write_text(txt, encoding="utf-8")
print("renamed audit helpers")
print(txt.count("auditItoa"), txt.count("auditAtoi"))
