package mcpv2

import (
	"bytes"
	"context"
	"encoding/json"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"go.uber.org/zap"
)

// SchemaValidator — DI & Testing ke liye interface (RULE 8-B:32)
type SchemaValidator interface {
	Validate(ctx context.Context, toolName string, args map[string]any) error
	Invalidate(toolName string)
}

// CachedValidator — Single Source of Truth = DB mcp_v2_tools.input_schema
// Service.ListTools/GetTool already Redis 5m cache karta hai, wahi se schema laayega.
// Ek baar compile -> sync.Map me lifetime cache. CreateTool par Invalidate hoga.
type CachedValidator struct {
	svc    *Service
	logger *zap.Logger
	cache  sync.Map // map[string]*jsonschema.Schema
}

func NewCachedValidator(svc *Service, logger *zap.Logger) *CachedValidator {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &CachedValidator{svc: svc, logger: logger}
}

func (v *CachedValidator) Invalidate(toolName string) {
	v.cache.Delete(toolName)
}

func (v *CachedValidator) Validate(ctx context.Context, toolName string, args map[string]any) error {
	if v.svc == nil || toolName == "" {
		return nil // fail-open for tests / unknown tool -> handler ka "tool not found" handle karega
	}
	if args == nil {
		args = map[string]any{}
	}
	compiled, err := v.getCompiled(ctx, toolName)
	if err != nil {
		v.logger.Warn("schema compile failed — fail-open", zap.String("tool", toolName), zap.Error(err))
		return nil // degrade gracefully, Service validation backup hai
	}
	if compiled == nil {
		return nil // no schema or tool not found -> allow, handler will return -32601
	}
	if err := compiled.Validate(args); err != nil {
		// jsonschema error already contains path: e.g. "/question: length must be >=1, but got 0"
		return NewInvalidInput("Invalid params: "+err.Error(), nil)
	}
	return nil
}

func (v *CachedValidator) getCompiled(ctx context.Context, toolName string) (*jsonschema.Schema, error) {
	if val, ok := v.cache.Load(toolName); ok {
		return val.(*jsonschema.Schema), nil
	}
	// Single Source of Truth — DB se lao. ListTools ka Redis cache hit = 0 DB call for valid tools second time se
	// GetTool use kar rahe hai, pehli baar DB hit, uske baad sync.Map hit = <0.1ms
	tool, err := v.svc.GetTool(ctx, toolName)
	if err != nil || tool.InputSchema == nil {
		return nil, nil // tool not found -> handler dynamic fallback will handle
	}
	
	schemaBytes, err := json.Marshal(tool.InputSchema)
	if err != nil {
		return nil, err
	}
	
	c := jsonschema.NewCompiler()
	if err := c.AddResource("schema.json", bytes.NewReader(schemaBytes)); err != nil {
		return nil, err
	}
	compiled, err := c.Compile("schema.json")
	if err != nil {
		return nil, err
	}
	
	v.cache.Store(toolName, compiled)
	return compiled, nil
}
