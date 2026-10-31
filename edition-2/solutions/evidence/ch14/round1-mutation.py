# Restore the independently observed EOF behavior in an isolated current-source copy.
import pathlib, shutil, tempfile, subprocess, json
base=pathlib.Path('/Users/bill/projects/ensemble/solutions/edition-2')
root=pathlib.Path(tempfile.mkdtemp(prefix='ensemble-ch14-eof-mutant-'))
try:
 tree=root/'main';shutil.copytree(base/'main',tree,ignore=shutil.ignore_patterns('node_modules','*.log','cr','solutions'))
 (tree/'gui/node_modules').symlink_to(base/'main/gui/node_modules',target_is_directory=True)
 p=tree/'gui/virtual-user.mjs';source=p.read_text();assert 'lines.once("close", close);' in source
 p.write_text(source.replace('lines.once("close", close);','lines.once("close", () => user.browser.close());'))
 run=subprocess.run(['node','--test','virtual-user.test.mjs'],cwd=tree/'gui',capture_output=True,text=True,timeout=30)
 result={'mutation':'restore browser-only EOF cleanup, leaving pending timer alive','applied':True,'exit':run.returncode,'stdout':run.stdout,'stderr':run.stderr}
 (base/'evidence/ch14/round1-mutation.json').write_text(json.dumps(result,indent=2));print(run.stdout,run.stderr)
finally:shutil.rmtree(root)
