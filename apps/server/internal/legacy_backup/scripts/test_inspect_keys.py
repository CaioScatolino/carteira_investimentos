import urllib.request
import urllib.parse
import json

headers = {
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36',
    'Referer': 'https://statusinvest.com.br/acoes/busca-avancada',
    'Accept': '*/*',
    'X-Requested-With': 'XMLHttpRequest',
    'Content-Type': 'application/x-www-form-urlencoded; charset=UTF-8'
}

for cat_type, name in [(1, "Acoes"), (2, "FIIs")]:
    url = f"https://statusinvest.com.br/category/advancedsearchresultpaginated?CategoryType={cat_type}"
    payload = {'search': '{}', 'CategoryType': cat_type, 'page': 0, 'take': 1000}
    data = urllib.parse.urlencode(payload).encode('utf-8')
    req = urllib.request.Request(url, data=data, headers=headers)
    with urllib.request.urlopen(req, timeout=10) as resp:
        content = resp.read()
        js = json.loads(content)
        items = js.get('list', [])
        print(f"=== {name}: {len(items)} items ===")
        # Print fields with sample non-null values
        sample = {}
        for item in items:
            for k, v in item.items():
                if k not in sample or (sample[k] is None and v is not None):
                    sample[k] = v
        for k, v in sorted(sample.items()):
            t = type(v).__name__ if v is not None else "null"
            print(f"  {k}: {t} (ex: {repr(v)})")
