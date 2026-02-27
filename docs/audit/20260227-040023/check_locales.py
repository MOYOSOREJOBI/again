import json,glob,os
files=glob.glob('web/messages/*.json')+glob.glob('web/locales/*.json')
print('locale_files',len(files))
if not files:
    print('no locale json files found')
    raise SystemExit(0)
allk=set();data={}
for f in files:
    loc=os.path.splitext(os.path.basename(f))[0]
    d=json.load(open(f))
    data[loc]=set(d.keys()); allk|=set(d.keys())
for loc,ks in sorted(data.items()):
    miss=sorted(allk-ks)
    print(loc,'missing',len(miss))
    if miss: print(' sample',miss[:10])
