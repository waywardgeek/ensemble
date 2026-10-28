// Temporary read-only entry-audit driver; not an adapter or acceptance test.
package main

import (
 "context"
 "encoding/json"
 "fmt"
 "net"
 "os"
 "os/exec"
 "path/filepath"
 "time"

 "github.com/gorilla/websocket"
 "github.com/waywardgeek/ensemble/internal/fakevendor"
 "github.com/waywardgeek/ensemble/internal/grade"
 "github.com/waywardgeek/ensemble/internal/gradeedition2"
)

func main() {
 if err := run(); err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
}
func run() error {
 if len(os.Args)!=4 { return fmt.Errorf("usage: audit baseline|pair ROOT EVIDENCE") }
 root, evidence := os.Args[2], os.Args[3]
 core := filepath.Join(root,"solutions/ch11")
 gui := filepath.Join(core,"gui")
 switch os.Args[1] {
 case "baseline":
  bin, cleanup, err := grade.Build(core); if err != nil {return err}; defer cleanup()
  example, err := gradeedition2.ReadConsumerCase(""); if err != nil{return err}
  media, err := gradeedition2.ReadMediaModels(""); if err != nil{return err}
  inputs:=map[string]any{"bin":bin,"core":core,"gui":gui,"gui_model":"claude-sonnet-5","consumer":filepath.Join(core,"ch05"),"workflow":filepath.Join(core,"ch06"),"exercise":filepath.Join(core,"ch07"),"stream_model":"claude-sonnet-5","consumer_case":example,"media":media}
  writeJSON(filepath.Join(evidence,"baseline-inputs.json"),inputs)
  report, err := gradeedition2.Chapter11(bin, core, gui, "claude-sonnet-5", filepath.Join(core,"ch05"),filepath.Join(core,"ch06"),filepath.Join(core,"ch07"),"claude-sonnet-5",example,media)
  if err!=nil{return err}
  return json.NewEncoder(os.Stdout).Encode(report)
 case "pair":
  bin, cleanup, err := grade.Build(gui); if err!=nil{return err}; defer cleanup()
  var outcomes []map[string]any
  for _, declared := range []bool{false,true} {
   name:="pair-empty"; if declared{name="pair-declared"}
   result:=map[string]any{"name":name,"declared_websocket_skill":declared}
   err:=probe(bin,filepath.Join(gui,"web/gui"),filepath.Join(evidence,name),declared,result)
   if err!=nil{result["error"]=err.Error()}
   outcomes=append(outcomes,result)
  }
  writeJSON(filepath.Join(evidence,"pair-results.json"),outcomes)
  return json.NewEncoder(os.Stdout).Encode(outcomes)
 }
 return fmt.Errorf("unknown mode")
}
func writeJSON(path string, v any) error { b,e:=json.MarshalIndent(v,"","  ");if e!=nil{return e};return os.WriteFile(path,append(b,'\n'),0600) }
func probe(bin, assets, work string, declared bool, result map[string]any) (retErr error) {
 os.WriteFile(filepath.Join(work,"boundary.txt"),[]byte("entry-audit-boundary\n"),0600)
 skills,primary:="",""
 if declared {
  skills=filepath.Join(work,"skills"); primary="browser"
  os.MkdirAll(filepath.Join(skills,primary),0700)
  text:="---\nname: browser\ndescription: Local audit browser boundary.\ntools: [read_file, browser_probe]\nmcp_servers:\n  - name: browser\n    transport: websocket\n---\nUse the browser boundary only when connected.\n"
  os.WriteFile(filepath.Join(skills,primary,"SKILL.md"),[]byte(text),0600)
 }
 fake:=fakevendor.New([]fakevendor.Reply{{Text:"Boundary observation complete."}})
 defer func(){writeJSON(filepath.Join(work,"fake-requests.json"),fake.Requests());fake.Close()}()
 l,e:=net.Listen("tcp","127.0.0.1:0");if e!=nil{return e}; port:=l.Addr().(*net.TCPAddr).Port;l.Close()
 ctx,cancel:=context.WithTimeout(context.Background(),18*time.Second);defer cancel()
 cmd:=exec.CommandContext(ctx,bin,"--port",fmt.Sprint(port),"--gui-dir",assets,"--model","claude-sonnet-5","--workspace",work,"--save",filepath.Join(work,"save.json"),"--event-log",filepath.Join(work,"events.jsonl"))
 cmd.Dir=work
 cmd.Env=append(os.Environ(),"LLM_VENDOR=anthropic","LLM_MODEL=claude-sonnet-5","LLM_API_KEY=fake-entry-audit-key","LLM_BASE_URL="+fake.URL()+"/anthropic","EN_DISABLE_STREAMING=1","EN_SKILLS_DIR="+skills,"EN_PRIMARY_SKILL="+primary)
 writeJSON(filepath.Join(work,"launch.json"),map[string]any{"args":cmd.Args,"cwd":work,"environment":cmd.Env})
 out,e:=os.Create(filepath.Join(work,"process.log"));if e!=nil{return e};defer out.Close();cmd.Stdout=out;cmd.Stderr=out
 stdin,e:=cmd.StdinPipe();if e!=nil{return e}
 if e=cmd.Start();e!=nil{return e}
 done:=make(chan error,1);go func(){done<-cmd.Wait()}()
 defer func(){stdin.Close();select{case err:=<-done:result["process_exit"]=fmt.Sprint(err);case <-time.After(3*time.Second):cancel();result["process_exit"]=fmt.Sprint(<-done)}}()
 var ws *websocket.Conn
 deadline:=time.Now().Add(5*time.Second)
 for time.Now().Before(deadline) {ws,_,e=websocket.DefaultDialer.Dial(fmt.Sprintf("ws://127.0.0.1:%d/ws",port),nil);if e==nil{break};time.Sleep(30*time.Millisecond)}
 if e!=nil{return e};defer ws.Close()
 transcript,e:=os.Create(filepath.Join(work,"websocket.jsonl"));if e!=nil{return e};defer transcript.Close(); enc:=json.NewEncoder(transcript)
 send:=func(m map[string]any) error {enc.Encode(map[string]any{"direction":"send","message":m});return ws.WriteJSON(m)}
 send(map[string]any{"type":"subscribe"})
 var methods []string
 initialized,listed,prompted,reverse:=false,false,false,false
 deadline=time.Now().Add(7*time.Second);ws.SetReadDeadline(deadline)
 for time.Now().Before(deadline) {
  var msg map[string]any
  if e=ws.ReadJSON(&msg);e!=nil{result["read_end"]=e.Error();break}
  enc.Encode(map[string]any{"direction":"receive","message":msg})
  if msg["type"]!="jsonrpc"{continue}
  rpc,_:=msg["payload"].(map[string]any);method,_:=rpc["method"].(string)
  if method!=""{methods=append(methods,method)}
  response:=map[string]any{"jsonrpc":"2.0","id":rpc["id"]}
  switch method {
  case "initialize":
   initialized=true;response["result"]=map[string]any{"protocolVersion":"2024-11-05","serverInfo":map[string]any{"name":"entry-audit","version":"1"},"capabilities":map[string]any{"tools":map[string]any{}}};send(map[string]any{"type":"jsonrpc","payload":response})
  case "tools/list":
   listed=true;response["result"]=map[string]any{"tools":[]any{map[string]any{"name":"browser_probe","description":"No native effect; audit declaration only.","inputSchema":map[string]any{"type":"object","properties":map[string]any{}}}}};send(map[string]any{"type":"jsonrpc","payload":response})
   // The reverse call follows tools/list on the same ordered transport.
   send(map[string]any{"type":"jsonrpc","payload":map[string]any{"jsonrpc":"2.0","id":"entry-reverse","method":"tools/call","params":map[string]any{"name":"read_file","arguments":map[string]any{"path":"boundary.txt"}}}})
  default:
   if rpc["id"]=="entry-reverse"{reverse=true;result["reverse_response"]=rpc;if !prompted{prompted=true;send(map[string]any{"type":"user_input","text":"Report boundary readiness."})}}
  }
  if initialized&&listed&&reverse&&len(fake.Requests())>0{break}
 }
 result["methods"]=methods;result["initialize_observed"]=initialized;result["tools_list_observed"]=listed;result["reverse_reply_observed"]=reverse
 requests:=fake.Requests();result["provider_requests"]=len(requests)
 var declarations []any
 for _,request:=range requests{var body map[string]any;if json.Unmarshal(request.Body,&body)==nil{if tools,ok:=body["tools"].([]any);ok{for _,raw:=range tools{tool,_:=raw.(map[string]any);declarations=append(declarations,tool["name"])}}}}
 result["provider_tool_names"]=declarations
 return nil
}
