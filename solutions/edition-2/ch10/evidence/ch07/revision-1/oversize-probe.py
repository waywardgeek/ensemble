"""Local bounded-wire regression; original and revised executables stay distinct."""
import hashlib,json,os,subprocess,sys,tempfile,time
from pathlib import Path
import websocket
binary=Path(sys.argv[1]).resolve();results=[]
with tempfile.TemporaryDirectory(prefix='ch07-oversize-') as temporary:
 env=os.environ.copy();env.update(LLM_VENDOR='anthropic',LLM_MODEL='fixture',LLM_API_KEY='fixture',LLM_BASE_URL='http://127.0.0.1:1',CH02_LOG=str(Path(temporary)/'session.log'))
 process=subprocess.Popen([str(binary),'--port','0'],cwd=temporary,env=env,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
 try:
  origin=process.stdout.readline().strip();assert origin.startswith('http://127.0.0.1:')
  for size,fragmented in [(65536,False),(65537,False),(131072,False),(1048576,False),(131072,True)]:
   ws=websocket.create_connection(origin.replace('http:','ws:')+'/ws',origin=origin,timeout=5)
   payload='{"type":"subscribe","id":"boundary"}';payload+=' '*(size-len(payload))
   sent_error=None
   try:
    if fragmented:
     for i in range(0,len(payload),4096):
      opcode=websocket.ABNF.OPCODE_TEXT if i==0 else websocket.ABNF.OPCODE_CONT
      ws.send_frame(websocket.ABNF.create_frame(payload[i:i+4096],opcode,fin=i+4096>=len(payload)))
    else:ws.send(payload)
   except OSError as error:sent_error=type(error).__name__
   frame=ws.recv_frame()
   if size==65536:
    message=json.loads(frame.data);ok=frame.opcode==1 and message['type']=='snapshot_begin';details=message['type']
   else:
    details=frame.data[2:].decode() if frame.opcode==8 else str(frame.opcode);ok=frame.opcode==8 and details=='message exceeds 65536 bytes'
   results.append({'size':size,'fragmented':fragmented,'send_error':sent_error,'ok':ok,'reply':details});ws.close()
 finally:
  process.terminate();process.wait(timeout=10)
 log=Path(temporary)/'session.log';calls=sum('"type":"request_sent"' in line for line in log.read_text().splitlines())
report={'binary':str(binary),'sha256':hashlib.sha256(binary.read_bytes()).hexdigest(),'model_requests':calls,'results':results}
Path(sys.argv[2]).write_text(json.dumps(report,indent=2)+'\n');print(json.dumps(report,indent=2));assert calls==0 and all(r['ok'] for r in results)
