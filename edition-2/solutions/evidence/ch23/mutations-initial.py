import pathlib,tempfile,subprocess,shutil,json
repo=pathlib.Path('/Users/bill/projects/ensemble')
evidence=repo/'solutions/edition-2/evidence/ch23'
work=pathlib.Path(tempfile.mkdtemp(prefix='ensemble-ch23-mutants-'))
subprocess.run(['git','archive','fcb98a9459b328314f0a7ff06efbb5e7c85fe07e:solutions/edition-2/main'],cwd=repo,stdout=open(work/'source.tar','wb'),check=True)
subprocess.run(['tar','-xf',str(work/'source.tar'),'-C',str(work)],check=True)
(work/'source.tar').unlink()
cases=[
 ('file-read-absolute','internal/sandbox/files.go','func (s *sandbox) ReadFile(path string) ([]byte, error) {','func (s *sandbox) ReadFile(path string) ([]byte, error) {\n if filepath.IsAbs(path) { return os.ReadFile(path) }','TestSandboxFiles'),
 ('file-write-absolute','internal/sandbox/files.go','func (s *sandbox) OpenFile(path string, flags int, mode fs.FileMode) (*os.File, error) {','func (s *sandbox) OpenFile(path string, flags int, mode fs.FileMode) (*os.File, error) {\n if filepath.IsAbs(path) { return os.OpenFile(path,flags,mode) }','TestSandboxFiles'),
 ('file-boundary-disabled','internal/sandbox/files.go','root := s.parent.Spec().SandboxRoot','root := ""','TestSandboxFiles'),
 ('host-path-error','internal/sandbox/files.go','return "", refusal(path)\n\t}\n\treturn resolved, nil','return "", fmt.Errorf("sandbox root %s refused %s",root,path)\n\t}\n\treturn resolved, nil','TestSandboxFiles'),
 ('kernel-read','internal/sandbox/command.go','(deny file-read*)','; read denial removed','TestSandboxCommands'),
 ('kernel-write','internal/sandbox/command.go','(deny file-write*)','; write denial removed','TestSandboxCommands'),
 ('network','internal/sandbox/command.go','(deny network*)','; network denial removed','TestSandboxCommands'),
 ('environment','internal/sandbox/command.go','out = append(out, "TERM=dumb")','out = append(out, "TERM=dumb", "ODD_CREDENTIAL="+os.Getenv("ODD_CREDENTIAL"))','TestSandboxCommands'),
 ('home','internal/sandbox/command.go','"HOME="+root','"HOME=/not-the-workspace"','TestSandboxCommands'),
 ('git-config','internal/sandbox/command.go','"GIT_CONFIG_GLOBAL="+filepath.Join(root, ".ensemble-gitconfig")','"GIT_CONFIG_GLOBAL="+os.Getenv("GIT_CONFIG_GLOBAL")','TestSandboxCommands'),
 ('child-derive','ensemble/spec.go','s.DataDir, s.LogPath, s.SavePath = dataDir, "", ""','s = AgentSpec{DataDir:dataDir}','TestSandboxChildSpec'),
 ('child-safe','ensemble/spec.go','s.SafeMode = true','s.SafeMode = false','TestSandboxChildSpec'),
 ('child-web','ensemble/spec.go','s.EnableWebSearch = false','s.EnableWebSearch = true','TestSandboxChildSpec'),
 ('child-root','ensemble/spec.go','s.SandboxRoot = parent.SandboxRoot','s.SandboxRoot = ""','TestSandboxChildSpec'),
 ('child-alarm','ensemble/spec.go','return s, fmt.Errorf("child permissions narrowed: %s", strings.Join(widened, ", "))','_ = fmt.Sprint(strings.Join(widened, ", ")); return s, nil','TestSandboxChildSpec'),
 ('safe-mode','internal/agent/agent.go','if a.spec.SafeMode {','if false {','TestSandboxSafeMode'),
 ('web-gate','internal/skills/skills.go','s.name == "web-search" && !parent.Spec().EnableWebSearch','s.name == "web-search" && false','TestSandboxWebGate'),
 ('diagnostic-redaction','ensemble/ensemble.go','e.logger.Print(e.sanitize(fmt.Sprintf(format, args...)))','e.logger.Print(fmt.Sprintf(format,args...))','TestSandboxCredentialLogs'),
 ('response-redaction','internal/engine/engine.go','if secret != "" {','if false {','TestSandboxCredentialLogs'),
]
results=[]
try:
 baseline=subprocess.run(['go','test','./ensemble','-run','TestSandbox','-count=1'],cwd=work,capture_output=True,text=True)
 (evidence/'mutation-baseline.txt').write_text(baseline.stdout+baseline.stderr)
 if baseline.returncode:raise RuntimeError('baseline failed')
 for name,file,old,new,test in cases:
  path=work/file;original=path.read_text();count=original.count(old)
  if count!=1:raise RuntimeError((name,'replacement count',count))
  path.write_text(original.replace(old,new))
  try:r=subprocess.run(['go','test','./ensemble','-run','^'+test+'$','-count=1','-v'],cwd=work,capture_output=True,text=True,timeout=30)
  finally:path.write_text(original)
  result={'name':name,'file':file,'old':old,'new':new,'test':test,'exit':r.returncode,'output':r.stdout+r.stderr}
  results.append(result)
  (evidence/'mutations.json').write_text(json.dumps(results,indent=2)+'\n')
  print(name,r.returncode,flush=True)
finally:shutil.rmtree(work)
