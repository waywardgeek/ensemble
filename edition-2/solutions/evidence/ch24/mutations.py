"""Fixed chapter-24 mutations, applied one at a time to disposable student source."""
import json, os, shutil, subprocess, tempfile, time
from pathlib import Path

root = Path(__file__).resolve().parents[4]
main = root / 'solutions/edition-2/main'
evidence = Path(__file__).resolve().parent
scratch = Path(tempfile.mkdtemp(prefix='ensemble-ch24-mutations-'))
work = scratch / 'main'
shutil.copytree(main, work, ignore=shutil.ignore_patterns('node_modules', '*.log', 'cr'))
os.symlink(main / 'gui/node_modules', work / 'gui/node_modules')
command = ['go','test','-run','TestExternalReuse','-count=1','-timeout=60s','.']
cases = [
 ('embedded-default', 'gui/gui.go', 'assets := http.FS(Assets())', 'assets := http.FileSystem(http.Dir("."))', command, 'embedded default returned no application shell'),
 ('disk-override', 'gui/gui.go', 'if directory != "" {', 'if false {', command, 'disk override did not select'),
 ('asset-root', 'gui/assets.go', 'fs.Sub(assets, "web/gui")', 'fs.Sub(assets, "web")', command, 'embedded index.html'),
 ('command-relocation', 'gui/cmd/main.go', 'flag.String("gui-dir", "",', 'flag.String("gui-dir", "web/gui",', ['go','test','-run','TestSavedBudgetEntry','-count=1','.'], 'relocated asset /: status=404'),
 ('component-shell-coupling', 'gui/web/gui/artifact-scroll.js', 'import { markdown,', 'import "./app.js";\nimport { markdown,', command, 'Failed to fetch dynamically imported module'),
 ('component-optional-speech', 'gui/web/gui/artifact-scroll.js', 'this.speech?.chunk', 'this.speech.chunk', command, "Cannot read properties of null"),
 ('application-speech-owner', 'gui/web/gui/app.js', 'document.querySelector("#artifacts"), speech)', 'document.querySelector("#artifacts"))', ['node','--test','responses.test.mjs'], 'Timeout'),
 ('public-replay', 'gui/gui.go', 'if e.Message.Actor == ensemble.Human {', 'if false {', command, 'public planted event absent'),
 ('public-prompt', 'gui/gui.go', 's.agent.Post(ensemble.UserMessage{Text: input.Text})', '_ = input.Text', command, 'i/o timeout'),
 ('relay-no-browser-id', 'gui/relay.go', '"id": id,', '"id": 0,', command, 'no-browser correlated error'),
 ('relay-restore-id', 'gui/mcp.go', 'response["id"] = original', 'response["id"] = request["id"]', command, 'relay correlation'),
 ('relay-unique-id', 'gui/mcp.go', 't.relayNext++', 't.relayNext = 1', command, 'relay IDs collide'),
 ('relay-response-routing', 'gui/gui.go', 'if tunnel.relayReply(input.Payload) {', 'if false && tunnel.relayReply(input.Payload) {', command, 'i/o timeout'),
 ('relay-browser-disconnect', 'gui/gui.go', 'close(tunnel.relayDone)', '_ = tunnel.relayDone', command, 'i/o timeout'),
 ('relay-listener-close', 'gui/gui.go', 'listener.Close()', '_ = listener', command, 'relay listener survived close'),
 ('relay-client-close', 'gui/gui.go', 'for conn := range s.relayConnections {\n\t\tconn.Close()', 'for conn := range s.relayConnections {\n\t\t_ = conn', command, 'relay shutdown stuck'),
 ('headless-mcp-import', 'internal/mcp/mcp.go', 'import (', 'import (\n _ "github.com/gorilla/websocket"', command, 'dependency control ./headless'),
 ('headless-command-import', 'main.go', 'import (', 'import (\n _ "github.com/gorilla/websocket"', command, 'dependency control ensemble:'),
]
selection = os.environ.get('CH24_MUTATIONS','')
if selection: cases = [case for case in cases if case[0] in selection.split(',')]
suffix = '-followup' if selection else ''
results=[]
try:
 baseline=subprocess.run(command,cwd=work/'gui',capture_output=True,text=True,timeout=75)
 (evidence/('mutation-baseline'+suffix+'.txt')).write_text(baseline.stdout+baseline.stderr)
 assert baseline.returncode==0, 'baseline failed'
 for name,filename,before,after,args,intended in cases:
  target=work/filename;original=target.read_text();assert before in original,name
  target.write_text(original.replace(before,after))
  originals={}
  if name.startswith('headless-'):
   for file in ['go.mod','go.sum']: originals[file]=(work/file).read_text()
   (work/'go.mod').write_text(originals['go.mod']+'\nrequire github.com/gorilla/websocket v1.5.3\n')
   (work/'go.sum').write_text(originals['go.sum']+(work/'gui/go.sum').read_text())
  started=time.monotonic()
  run=subprocess.run(args,cwd=work/'gui',capture_output=True,text=True,timeout=75)
  output=run.stdout+run.stderr
  (evidence/('mutation-'+name+suffix+'.txt')).write_text(output)
  item={'name':name,'file':filename,'before':before,'after':after,'command':args,'exit':run.returncode,
    'seconds':round(time.monotonic()-started,3),'intended_failure':intended,
    'observed_intended_failure':run.returncode!=0 and intended in output}
  results.append(item)
  (evidence/('mutations'+suffix+'.json')).write_text(json.dumps(results,indent=2)+'\n')
  print(name,run.returncode,item['observed_intended_failure'],flush=True)
  target.write_text(original)
  for file,data in originals.items():(work/file).write_text(data)
finally:
 shutil.rmtree(scratch)
