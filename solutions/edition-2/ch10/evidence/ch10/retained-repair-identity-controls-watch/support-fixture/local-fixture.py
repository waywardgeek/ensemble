"""Explicit LOCAL fixture for later binary integration. Never provider evidence.

No outbound network or credentials. At most 24 requests; each response supplies
synthetic usage and a literal LOCAL FIXTURE label. Select vendor from the path.
"""
import argparse
import http.server
import json
import re
import threading


def response(path, body, number):
    vendor='gemini' if '/models/' in path else ('openai' if '/chat/completions' in path else 'anthropic')
    messages=body.get('messages',body.get('contents',[]));last=messages[-1] if messages else {}
    text=json.dumps(last,ensure_ascii=False)
    tool_result=(last.get('role')=='tool' or 'tool_result' in text or 'functionResponse' in text)
    filenames=('marker.txt','policy-proof.txt','skill-proof.txt','after-resume.txt')
    selected=next((f for f in filenames if 'create '+f in text),None)
    marker=re.search(r'CH10-[A-D]-(?:anthropic|openai|gemini)',text)
    call=selected is not None and not tool_result
    args={'path':selected,'content':(marker.group(0) if marker else 'LOCAL-FIXTURE')+'\n','overwrite':False,'append':False}
    answer='LOCAL FIXTURE only. '+(marker.group(0) if marker else 'synthetic continuation')
    model=body.get('model','fixture');call_id='local-fixture-'+str(number)
    if vendor=='anthropic':
        part={'type':'tool_use','id':call_id,'name':'write_file','input':args} if call else {'type':'text','text':answer}
        normal={'model':model,'content':[part],'stop_reason':'tool_use' if call else 'end_turn','usage':{'input_tokens':1,'output_tokens':1}}
        stream=[{'type':'message_start','message':{'model':model,'usage':{'input_tokens':1,'output_tokens':0}}},{'type':'content_block_start','index':0,'content_block':part},{'type':'content_block_stop','index':0},{'type':'message_delta','delta':{'stop_reason':normal['stop_reason']},'usage':{'output_tokens':1}},{'type':'message_stop'}]
    elif vendor=='openai':
        msg={'role':'assistant','tool_calls':[{'id':call_id,'type':'function','function':{'name':'write_file','arguments':json.dumps(args)}}]} if call else {'role':'assistant','content':answer}
        normal={'model':model,'choices':[{'message':msg,'finish_reason':'tool_calls' if call else 'stop'}],'usage':{'prompt_tokens':1,'completion_tokens':1}}
        delta=json.loads(json.dumps(msg))
        if call:delta['tool_calls'][0]['index']=0
        stream=[{'model':model,'choices':[{'index':0,'delta':delta,'finish_reason':None}]},{'model':model,'choices':[{'index':0,'delta':{},'finish_reason':normal['choices'][0]['finish_reason']}]},{'choices':[],'usage':normal['usage']}]
    else:
        part={'functionCall':{'id':call_id,'name':'write_file','args':args}} if call else {'text':answer}
        normal={'candidates':[{'index':0,'content':{'role':'model','parts':[part]},'finishReason':'STOP'}],'usageMetadata':{'promptTokenCount':1,'candidatesTokenCount':1}}
        stream=[normal]
    if body.get('stream') or ':streamGenerateContent' in path:
        data=''.join('data: '+json.dumps(x)+'\n\n' for x in stream)
        if vendor=='openai':data+='data: [DONE]\n\n'
        return 'text/event-stream',data.encode()
    return 'application/json',json.dumps(normal).encode()


class Server(http.server.ThreadingHTTPServer):
    def __init__(self):
        self.count=0;self.lock=threading.Lock();super().__init__(('127.0.0.1',0),Handler)
class Handler(http.server.BaseHTTPRequestHandler):
    def log_message(self,*_):pass
    def do_POST(self):
        with self.server.lock:self.server.count+=1;number=self.server.count
        if number>24:self.send_error(429,'local fixture ceiling');return
        length=int(self.headers.get('content-length','-1'))
        if not 0<=length<=2*1024*1024:self.send_error(400);return
        self.connection.settimeout(5);body=json.loads(self.rfile.read(length))
        content,data=response(self.path,body,number)
        self.send_response(200);self.send_header('Content-Type',content);self.send_header('Content-Length',str(len(data)));self.end_headers();self.wfile.write(data)


if __name__=='__main__':
    parser=argparse.ArgumentParser();parser.add_argument('--seconds',type=int,default=600);args=parser.parse_args()
    if not 1<=args.seconds<=600:parser.error('deadline must be 1..600 seconds')
    server=Server();timer=threading.Timer(args.seconds,server.shutdown);timer.start()
    print(json.dumps({'source':'LOCAL FIXTURE, not a provider','origin':f'http://127.0.0.1:{server.server_port}','request_ceiling':24}),flush=True)
    try:server.serve_forever()
    except KeyboardInterrupt:pass
    finally:timer.cancel();server.server_close()
