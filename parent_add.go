// Phase 1.5 additive
type ChunkConfig struct {TargetSize int;MinSize int;MaxSize int;Overlap int;SoftLimit int;HardLimit int;OverlapTokens int;AtomicCodeFence bool;WordCountEstimate bool}
func (c ChunkConfig) effectiveHard() int {if c.HardLimit!=0{return c.HardLimit};if c.MaxSize!=0{return c.MaxSize};return c.SoftLimit+300}
func (c ChunkConfig) effectiveSoft() int {if c.SoftLimit!=0{return c.SoftLimit};if c.TargetSize!=0{return c.TargetSize};return 1200}
type ParentChunk struct {Text string;PageIndex int;SectionPath string;TokenCount int;ChunkHash string}
func DefaultParentConfig() ChunkConfig {return ChunkConfig{TargetSize:1200,MinSize:900,MaxSize:1500,Overlap:20,SoftLimit:1200,HardLimit:1500,OverlapTokens:20,AtomicCodeFence:true,WordCountEstimate:true}}
func DefaultChildConfig() ChunkConfig {return ChunkConfig{TargetSize:150,MinSize:100,MaxSize:220,Overlap:20,SoftLimit:150,HardLimit:220,OverlapTokens:20,AtomicCodeFence:true,WordCountEstimate:true}}
func estimateTokensWithConfig(text string, wordCount bool) int {if wordCount {if text==""{return 0};w:=len(strings.Fields(text));if w==0{return estimateTokens(text)};est:=int(float64(w)*1.3);if est<1{est=1};return est};return estimateTokens(text)}
func (c *TextChunker) ChunkMarkdown(ctx context.Context, raw string, parentCfg, childCfg ChunkConfig) ([]ParentChunk, []TextChunk) {
if ctx!=nil{select{case <-ctx.Done():return nil,nil;default:}}
raw=cleanText(raw);if raw==""{return nil,nil}
childChunkerCfg:=ChunkConfig{TargetSize:childCfg.effectiveSoft(),MinSize:childCfg.MinSize,MaxSize:childCfg.effectiveHard(),Overlap:childCfg.OverlapTokens}
if childChunkerCfg.MinSize==0{childChunkerCfg.MinSize=100}
if childChunkerCfg.TargetSize==0{childChunkerCfg.TargetSize=150}
childChunker:=&TextChunker{cfg: ChunkerConfig{TargetSize: childChunkerCfg.TargetSize, MinSize: childChunkerCfg.MinSize, MaxSize: childChunkerCfg.MaxSize, Overlap: childChunkerCfg.Overlap}, separators: c.separators}
if len(childChunker.separators)==0{childChunker.separators=[]string{"\n\n","\n",". ","! ","? "," "}}
children:=childChunker.Chunk(raw)
for i:=range children{if children[i].TokenCount==0{children[i].TokenCount=estimateTokensWithConfig(children[i].Text, childCfg.WordCountEstimate)};children[i].ParentIndex=-1}
parents:=buildParentsFromChildren(children, parentCfg)
return parents, children
}
func buildParentsFromChildren(children []TextChunk, cfg ChunkConfig) []ParentChunk {
pHard:=cfg.effectiveHard();pSoft:=cfg.effectiveSoft()
if pHard==0{pHard=1500}
if pSoft==0{pSoft=1200}
var parents []ParentChunk
var cur []string
curTokens:=0;curSection:="";pageIdx:=0
flush:=func(){if len(cur)==0{return};txt:=strings.TrimSpace(strings.Join(cur, "\n\n"));if txt==""{cur=nil;curTokens=0;return};parents=append(parents, ParentChunk{Text: txt, PageIndex: pageIdx, SectionPath: curSection, TokenCount: curTokens, ChunkHash: HashChunkText(txt)});pageIdx++;cur=nil;curTokens=0;curSection=""}
for _,ch:=range children{ct:=ch.TokenCount;if ct==0{ct=estimateTokensWithConfig(ch.Text, cfg.WordCountEstimate)};if curSection==""{curSection=ch.SectionPath};if curTokens+ct>pHard && curTokens>0{flush();curSection=ch.SectionPath};cur=append(cur, ch.Text);curTokens+=ct;if curTokens>=pSoft{flush()}}
flush()
return parents
}
