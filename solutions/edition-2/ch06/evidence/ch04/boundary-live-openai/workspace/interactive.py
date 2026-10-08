import time
print("READY>",flush=True)
while True:
    try: value=input()
    except EOFError: break
    time.sleep(0.2)
    print("REPLY:"+value+" READY>",flush=True)
