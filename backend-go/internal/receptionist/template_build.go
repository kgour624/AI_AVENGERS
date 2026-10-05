package receptionist

import (
 "context"
 "encoding/json"
 "fmt"
 "strings"
 "github.com/google/uuid"
 "go.uber.org/zap"
)

func (t *TemplateOrchestrator) BuildCheckpoints(ctx context.Context, agenda string, points []string, hits []string) [][]string {
 all := make([][]string, len(points))
 hitsStr := strings.Join(hits, " | ")
 for i, pt := range points {
  prompt := fmt.Sprintf("Agenda:%s | Point:%s | Hits:%s | Is point ke liye 4 specific Hinglish question bana, har line ek question, no numbering.", agenda, pt, hitsStr)
  raw, _ := t.search.Search(ctx, prompt) // reuse search as LLM fallback if gateway not direct
  _ = raw
  // try LLM if available
  qs := []string{}
  if t.llm != nil {
   if s, err := t.llm.Complete(ctx, "You are a receptionist bot helping with checklists", prompt); err==nil && s!="" {
    for _, l := range strings.Split(s, "\n") { l=strings.TrimSpace(l); if len(l)>6 { qs=append(qs,l) } }
   }
  }
  if len(qs)==0 { qs = t.fallbackPRD(pt) }
  all[i]=qs
  t.logger.Info("build checkpoints", zap.Int("point", i), zap.Int("qs", len(qs)))
 }
 return all
}
func (t *TemplateOrchestrator) fallbackPRD(pt string) []string {
 return []string{
  pt + " ka platform kya hoga - Mobile / Watch / Web?",
  pt + " me kaunse sensors lagenge - Heart, Sleep, Steps, BP?",
  pt + " ka data 24*7 kaha store hoga - local ya cloud?",
  pt + " ka battery vs real-time tradeoff kya rakhe?",
 }
}
var _ = json.Marshal
var _ = uuid.New
var _ = fmt.Sprint
