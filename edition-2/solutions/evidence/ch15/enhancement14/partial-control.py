exec(open('/tmp/ch14_mutations.py').read().split('results=[]')[0])
label='first-read-partial';file='internal/tools/tools.go'
with tempfile.TemporaryDirectory(prefix='ch14-partial-') as tmp:
 tree=pathlib.Path(tmp)/'source';shutil.copytree(source,tree)
 path=tree/file;original=path.read_text();mutated=original.replace('return result, nil','if strings.Contains(result, "FILE01-CONTENT-") { return strings.SplitAfter(result, "\\n")[0], nil }; return result, nil',1);path.write_text(mutated)
 out.joinpath(label+'.patch').write_text(''.join(difflib.unified_diff(original.splitlines(True),mutated.splitlines(True),fromfile=file,tofile=file)))
 r=subprocess.run(['/tmp/grade-edition2-context','-ch','14','-json','-gui',str(tree/'gui'),str(tree)],capture_output=True,text=True)
 name='first-read-partial-'+__import__('sys').argv[1]
 out.joinpath(name+'.json').write_text(r.stdout);out.joinpath(name+'.stderr').write_text(r.stderr);print(name,r.returncode,json.loads(r.stdout)['score'])
