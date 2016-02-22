import subprocess

out_bytes = subprocess.check_output(['ls','-a'])
out_text = out_bytes.decode('utf-8')

print(out_text)