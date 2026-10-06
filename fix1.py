import pathlib
p=pathlib.Path("backend-go/internal/receptionist/orchestrator.go")
t=p.read_text(encoding="utf-8")
old="    prompt := fmt.Sprintf(\"User message: '%s'. Acknowledge this briefly in 1-2 lines in %s as a %s. Say you noted it for this section.\", message, sess.Language, sess.Persona)"
new="    # FIXED: Mother-like Secretary with Fear/Pressure + Judge\n    pressure:=5\n    lang:=sess.Language\n    if lang==\"\" { lang=\"HINGLISH\" }\n    persona:=sess.Persona\n    if persona==\"\" { persona=\"software engineer\" }\n    agenda:=sess.Agenda\n    lower:=message\n    # intent\n    prompt:=\"\"\n    if len(message)>0 {\n        # mother spoon-feed one question\n        prompt = fmt.Sprintf(\"You are Brilliant Secretary fear revenue loss 50L if checkpoint fails pressure %d/100 language %s persona %s. Mother-like spoon-feed ONE question only for agenda '%s'. User said '%s'. Ask Haan/Nahi/Samjhao warm Hinglish.\", pressure, lang, persona, agenda, message)\n    }"
if old in t:
    t=t.replace(old,new)
    p.write_text(t,encoding="utf-8")
    print("fix1 done")
else:
    print("old not found fix1")
