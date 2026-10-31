"""Small source mutations in one disposable copy; original grader is untouched."""
import json, pathlib, shutil, subprocess, tempfile
HERE=pathlib.Path(__file__).resolve().parent
MAIN=HERE.parents[1]/'main'
results=[]
cases=[
 ('handshake-refusal','gateway/noise.go','return data, err\n}\n\n// IsHandshakeComplete','return data, nil\n}\n\n// IsHandshakeComplete','./gateway','TestNoiseKK','wrong static key authenticated'),
 ('ciphertext-authentication','gateway/noise.go','return n.receive.Decrypt(nil, associated, data)','plain,err:=n.receive.Decrypt(nil,associated,data);if err!=nil{return []byte("tampered"),nil};return plain,nil','./gateway','TestNoiseKK','tampered ciphertext accepted'),
 ('key-stability','gateway/noise.go','return Keypair{Private: k.Bytes(), Public: k.PublicKey().Bytes()}, nil\n}\n\n// NoiseKK','_ = k; return GenerateKeypair()\n}\n\n// NoiseKK','./gateway','TestIdentityFiles','identity changed'),
 ('provenance','internal/gateway/admin.go','Content: wire.ProvenanceHeader(source) + "\\n\\n" + message.Content','Content: message.Content','./internal/gateway','TestRoutingAuthenticatesAndStamps','bad provenance/envelope'),
 ('contactability','internal/gateway/admin.go','if !slices.Contains(s.state.Contactable[message.ToAgent], name) {','if name != "outsiders" && !slices.Contains(s.state.Contactable[message.ToAgent], name) {','./internal/gateway','TestRoutingAuthenticatesAndStamps','unauthorized send accepted'),
 ('offline-queue','internal/gateway/admin.go','q.Messages = append(q.Messages, env)','if s.active[message.ToAgent]!=nil{q.Messages = append(q.Messages, env)}','./internal/gateway','TestDeliveryAckResume','delivery timeout'),
 ('ack-after-handler','gateway/client.go','c.owner.OnDeliver(*frame.Envelope)','c.write(context.Background(),Frame{Type:"ack",Seq:frame.Envelope.Seq});c.ack.Store(frame.Envelope.Seq)\n c.owner.OnDeliver(*frame.Envelope)','./internal/gateway','TestDeliveryAckResume','ack preceded handler acceptance'),
 ('receipt-resume','internal/gateway/server.go','q.Ack = seq','return\n q.Ack = seq','./internal/gateway','TestDeliveryAckResume','offline resume'),
 ('narrow-grant','internal/gateway/admin.go','if !slices.Contains(s.memberships(id), g) {','if len(g)<0 {','./internal/gateway','TestChildrenAndPersistedPolicy','over-grant accepted'),
 ('admin-capability','internal/gateway/admin.go','if !s.capable(id, "group-admin") {','if false {','./internal/gateway','TestChildrenAndPersistedPolicy','non-admin edited groups'),
 ('config-immutability','internal/gateway/admin.go','if !validName(a.Name) || !isRuntime(a.Name) {','if !validName(a.Name) {','./internal/gateway','TestChildrenAndPersistedPolicy','config group mutated'),
 ('persisted-policy','internal/gateway/server.go','return writeJSON(filepath.Join(s.config.Gateway.DataDir, "state.json"), s.state)','return nil','./internal/gateway','TestChildrenAndPersistedPolicy','websocket: bad handshake'),
 ('actor-admission','cmd/gateway-agent/main.go','a.agent.Submit(a.ctx, e.Content)','a.agent.Submit(a.ctx, "lost gateway provenance")','./cmd/gateway-agent','TestGatewayUsesActualActor','provenance did not reach model'),
 ('response-correlation','gateway/client.go','ConversationID: request.ConversationID','ConversationID: "wrong-conversation"','./cmd/gateway-agent','TestGatewayUsesActualActor','wrong reply'),
]
cases=[case for case in cases if case[0] in ['key-stability','contactability','narrow-grant','persisted-policy']]
with tempfile.TemporaryDirectory(prefix='gateway26-mutations-') as scratch:
    root=pathlib.Path(scratch)/'main'
    shutil.copytree(MAIN,root,ignore=shutil.ignore_patterns('node_modules','.git','cr'))
    baseline=subprocess.run(['go','test','./gateway','./internal/gateway','./cmd/gateway-agent'],cwd=root,text=True,capture_output=True,timeout=30)
    results.append({'baseline':True,'exit':baseline.returncode,'output':baseline.stdout+baseline.stderr})
    if baseline.returncode:raise SystemExit('baseline failed')
    for name,path,old,new,pkg,test,expected in cases:
        target=root/path;original=target.read_text();count=original.count(old)
        if count!=1:
            results.append({'name':name,'applied':False,'matches':count});continue
        target.write_text(original.replace(old,new,1))
        try:
            p=subprocess.run(['go','test','-count=1','-timeout=12s',pkg,'-run','^'+test+'$'],cwd=root,text=True,capture_output=True,timeout=20)
            output=p.stdout+p.stderr
            results.append({'name':name,'file':path,'old':old,'new':new,'applied':True,'exit':p.returncode,'expected':expected,'intended_failure':p.returncode!=0 and expected in output,'output':output})
        finally:target.write_text(original)
    restored=subprocess.run(['go','test','-race','./gateway','./internal/gateway','./cmd/gateway-agent'],cwd=root,text=True,capture_output=True,timeout=30)
    results.append({'restored':True,'exit':restored.returncode,'output':restored.stdout+restored.stderr})
(HERE/'mutation-corrections.json').write_text(json.dumps(results,indent=2)+'\n')
print(json.dumps([{k:v for k,v in row.items() if k not in ['output','old','new']} for row in results],indent=2))
