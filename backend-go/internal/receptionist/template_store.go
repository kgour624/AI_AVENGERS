package receptionist

import (
 "context"
 "fmt"
 "strings"
)

type TemplateStore struct {
	deps *Deps
}

func (t *TemplateStore) BuildCheckpoints(ctx context.Context, agenda string, points []string, hits []string) [][]string {
 all := make([][]string, len(points))
 for i, pt := range points {
  prompt := fmt.Sprintf("Agenda:%s | Point:%s | Hits:%v | Is point ke liye 4 specific Hinglish question bana, har line ek question.", agenda, pt, strings.Join(hits, " "))
  raw, _ := t.deps.Gateway.Cheap(ctx, prompt, 0.6)
  qs := parseQuestions(raw)
  if len(qs)==0 { qs = t.fallbackPRD(pt) }
  all[i]=qs
 }
 return all
}

func (t *TemplateStore) fallbackPRD(pt string) []string {
 return []string{
  pt + " ka platform kya hoga - Mobile / Watch / Web?",
  pt + " me kaunse sensors - Heart, Sleep, Steps, BP?",
  pt + " ka data 24*7 kaha store hoga - local ya cloud?",
  pt + " ka battery vs real-time ka tradeoff kya rakhe?",
 }
}

func parseQuestions(s string) []string {
 lines:=strings.Split(s, "\n"); var out []string
 for _,l:=range lines{ l=strings.TrimSpace(l); if(len(l)>5){out=append(out,l)} }
 return out
}
