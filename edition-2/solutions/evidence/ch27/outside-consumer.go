// This source is copied into an unrelated module by validate.py. Only the public
// gateway package is available to the consumer; the actual binary owns routing.
package main

import (
 "context"
 "encoding/json"
 "fmt"
 "os"
 "os/exec"
 "path/filepath"
 "strings"
 "time"

 "ensemble/gateway"
)
type handler struct{gateway.NopHandler;in chan gateway.Envelope}
func(h *handler)OnDeliver(e gateway.Envelope){h.in<-e}
func main(){if err:=run();err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(1)}}
func run()error{
 binary,dir,address:=os.Args[1],os.Args[2],os.Args[3]
 parent,_:=gateway.GenerateKeypair();recipient,_:=gateway.GenerateKeypair();child,_:=gateway.GenerateKeypair()
 cfg:=map[string]any{"gateway":map[string]string{"listen":address,"data_dir":dir},"groups":map[string]any{"agents":map[string]any{"trust":"trusted","capabilities":[]string{"send-message","declare-contactable"}}},"agents":map[string]any{}}
 agents:=cfg["agents"].(map[string]any)
 for id,key:=range map[string]gateway.Keypair{"parent":parent,"recipient":recipient}{agents[id]=map[string]any{"public_key":gateway.EncodePublicKey(key.Public),"groups":[]string{"agents"}}}
 data,_:=json.Marshal(cfg);config:=filepath.Join(dir,"config.json");if err:=os.WriteFile(config,data,0600);err!=nil{return err}
 keyText,err:=exec.Command(binary,"-config",config,"-print-key").Output();if err!=nil{return err}
 peer,err:=gateway.ParsePublicKey(strings.TrimSpace(string(keyText)));if err!=nil{return err}
 cmd:=exec.Command(binary,"-config",config);cmd.Stdout=os.Stderr;cmd.Stderr=os.Stderr;if err=cmd.Start();err!=nil{return err}
 defer func(){cmd.Process.Signal(os.Interrupt);cmd.Wait()}()
 ctx,cancel:=context.WithTimeout(context.Background(),10*time.Second);defer cancel()
 connect:=func(id string,key gateway.Keypair,h gateway.Handler)(*gateway.Client,error){return gateway.Connect(ctx,gateway.ClientConfig{GatewayURL:"ws://"+address+"/v1/ws",AgentID:id,StaticKey:key,GatewayPubKey:peer},h)}
 var p *gateway.Client
 for {p,err=connect("parent",parent,gateway.NopHandler{});if err==nil{break};select{case <-ctx.Done():return err;case <-time.After(20*time.Millisecond):}}
 defer p.Close()
 h:=&handler{in:make(chan gateway.Envelope,2)};r,err:=connect("recipient",recipient,h);if err!=nil{return err};defer r.Close()
 if err=r.DeclareContactable(ctx,[]string{"agents"});err!=nil{return err}
 if err=p.RegisterChild(ctx,gateway.ChildRegistration{AgentID:"child",PublicKey:child.Public,Groups:[]string{"agents"}});err!=nil{return err}
 c,err:=connect("child",child,gateway.NopHandler{});if err!=nil{return err};defer c.Close()
 if err=c.Send(ctx,gateway.OutboundMessage{ToAgent:"recipient",Content:"outside module delivered",ConversationID:"external",ResponseMode:gateway.ResponseAsync});err!=nil{return err}
 select{case e:=<-h.in:
  if e.Source.UserID!="child" || e.Source.Trust!="trusted" || e.Content!="[GATEWAY source=agents user=child trust=trusted]\n\noutside module delivered"{return fmt.Errorf("unexpected envelope: %+v",e)}
 case <-ctx.Done():return ctx.Err()}
 fmt.Println("PASS: outside module authenticated, registered a child, and delivered stamped content through the real gateway binary")
 return nil
}
