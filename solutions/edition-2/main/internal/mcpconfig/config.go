// Package mcpconfig validates physical launch configuration without starting peers.
package mcpconfig

import (
 "encoding/json"
 "fmt"
 "io"
 "os"
 "path/filepath"
 "regexp"
 "example.com/ensemble/internal/common"
)

type Reader struct { parent common.Ensemble }
func New(parent common.Ensemble) *Reader { return &Reader{parent:parent} }
func (r *Reader) Read(path string) (common.MCPFileConfig,error) {
 var out common.MCPFileConfig
 path,err:=filepath.Abs(path);if err!=nil{return out,err}
 f,err:=os.Open(path);if err!=nil{return out,fmt.Errorf("cannot open MCP configuration")};defer f.Close()
 raw,err:=io.ReadAll(io.LimitReader(f,(1<<20)+1));if err!=nil{return out,fmt.Errorf("cannot read MCP configuration")}
 v,err:=r.parent.JSON().Parse(raw,common.JSONBounds{Bytes:1<<20,Depth:64,Nodes:100000,Collection:100000,Scalars:true});if err!=nil{return out,fmt.Errorf("invalid MCP configuration JSON")}
 bad:=fmt.Errorf("invalid MCP configuration fields")
 exact:=func(m map[string]any,keys ...string)bool{if len(m)!=len(keys){return false};for _,k:=range keys{if _,ok:=m[k];!ok{return false}};return true}
 root,ok:=v.(map[string]any);if !ok||!exact(root,"version","connections","bindings")||root["version"]!=json.Number("1"){return out,bad}
 conns,ok:=root["connections"].([]any);if !ok||len(conns)>32{return out,bad}; bindings,ok:=root["bindings"].([]any);if !ok||len(bindings)>1024{return out,bad}
 keyOK:=func(s string)bool{ok,_:=regexp.MatchString(`^[a-z][a-z0-9_]{0,63}$`,s);return ok}
 list:=func(v any)([]string,bool){a,ok:=v.([]any);ret:=[]string{};for _,x:=range a{s,yes:=x.(string);if !yes{return nil,false};ret=append(ret,s)};return ret,ok}
 keys:=map[string]bool{};aliases:=map[string]bool{}
 for _,v:=range conns{
  m,ok:=v.(map[string]any);if !ok||!exact(m,"key","transport","command","args","cwd","env_allowlist","call_timeout_seconds"){return out,bad}
  key,_:=m["key"].(string);command,_:=m["command"].(string);cwd,_:=m["cwd"].(string)
  args,okA:=list(m["args"]);env,okE:=list(m["env_allowlist"]);n,okN:=m["call_timeout_seconds"].(json.Number)
  timeout:=0;if okN{v,e:=n.Int64();if e==nil&&v>=1&&v<=600{timeout=int(v)}}
  if !keyOK(key)||keys[key]||m["transport"]!="stdio"||command==""||cwd==""||!okA||!okE||timeout==0{return out,bad}
  seen:=map[string]bool{};for _,name:=range env{valid,_:=regexp.MatchString(`^[A-Za-z_][A-Za-z0-9_]*$`,name);if !valid||seen[name]{return out,bad};seen[name]=true}
  if !filepath.IsAbs(command){command=filepath.Join(filepath.Dir(path),command)};if !filepath.IsAbs(cwd){cwd=filepath.Join(filepath.Dir(path),cwd)}
  out.Connections=append(out.Connections,common.MCPStdioSpec{Key:key,CallTimeoutSeconds:timeout,Options:common.MCPStdioOptions{Command:command,CWD:cwd,Args:args,EnvAllowlist:env}});keys[key]=true
 }
 for _,v:=range bindings{
  m,ok:=v.(map[string]any);if !ok||!exact(m,"alias","connection","remote_name"){return out,bad}
  a,_:=m["alias"].(string);key,_:=m["connection"].(string);name,_:=m["remote_name"].(string);valid,_:=regexp.MatchString(`^[A-Za-z0-9_.-]{1,128}$`,name)
  if !keyOK(a)||aliases[a]||!keys[key]||!valid{return out,bad};aliases[a]=true
  out.Bindings=append(out.Bindings,common.MCPSelection{Alias:a,Connection:key,RemoteName:name})
 }
 return out,nil
}
