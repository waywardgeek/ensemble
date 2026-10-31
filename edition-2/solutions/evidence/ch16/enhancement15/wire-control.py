import pathlib,tempfile,shutil,subprocess,os,json,difflib
root=pathlib.Path.cwd();out=root/'solutions/edition-2/evidence/ch16/enhancement15'
with tempfile.TemporaryDirectory(prefix='ch15-wire-control-') as tmp:
 tmp=pathlib.Path(tmp);tree=tmp/'source';shutil.copytree(root/'solutions/edition-2/ch15',tree)
 p=tree/'internal/engine/engine.go';old=p.read_text();needle='func (e *engine) Send(ctx context.Context, body []byte, cb common.StreamCallbacks) (events []common.Event, err error) {'
 addition='''
 var wire map[string]any; json.Unmarshal(body,&wire)
 if text,ok:=wire["system"].(string);ok { wire["system"]=[]any{map[string]any{"type":"text","text":text}} }
 if messages,ok:=wire["messages"].([]any);ok {for _,value:=range messages {m:=value.(map[string]any);if blocks,ok:=m["content"].([]any);ok && len(blocks)==1 {b:=blocks[0].(map[string]any);if b["type"]=="text" {m["content"]=b["text"]}}}}
 body,_=json.Marshal(wire)
'''
 assert old.count(needle)==1;new=old.replace(needle,needle+addition);p.write_text(new);(out/'valid-wire-forms.patch').write_text(''.join(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='internal/engine/engine.go',tofile='internal/engine/engine.go')))
 overlay=tmp/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(root/'internal/gradeedition2/ch15_test.go'):str(out/'probe.go.txt')}}))
 env=dict(os.environ,MEMORY_PROBE_GUI=str(tree/'gui'),MEMORY_PROBE_RESULT=str(out/'valid-wire-forms.json'))
 r=subprocess.run(['go','test','-overlay',str(overlay),'./internal/gradeedition2','-run','^TestMemoryProbe$','-count=1'],capture_output=True,text=True,env=env,timeout=90);(out/'valid-wire-forms.txt').write_text(r.stdout+r.stderr);print(r.returncode,r.stdout+r.stderr);assert r.returncode==0
