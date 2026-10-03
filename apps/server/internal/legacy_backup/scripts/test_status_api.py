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

# 1. Test POST /category/advancedsearchresultpaginated with take=1000
for cat_type in [1, 2]:
    name = "Ações" if cat_type == 1 else "FIIs"
    url = f"https://statusinvest.com.br/category/advancedsearchresultpaginated?CategoryType={cat_type}"
    payload = {
        'search': '{}',
        'CategoryType': cat_type,
        'page': 0,
        'take': 1000
    }
    try:
        data = urllib.parse.urlencode(payload).encode('utf-8')
        req = urllib.request.Request(url, data=data, headers=headers)
        with urllib.request.urlopen(req, timeout=10) as resp:
            content = resp.read()
            js = json.loads(content)
            print(f"=== {name} (Paginated take=1000) ===")
            if isinstance(js, dict):
                print(f"  Dict keys: {list(js.keys())}")
                list_items = js.get('list', js.get('data', []))
                print(f"  Total items returned: {len(list_items)}, total records: {js.get('total')}")
                if len(list_items) > 0:
                    print(f"  Sample keys ({len(list_items[0])} columns): {list(list_items[0].keys())}")
                    print(f"  Sample row: {list_items[0]}")
            elif isinstance(js, list):
                print(f"  List length: {len(js)}")
                if len(js) > 0:
                    print(f"  Sample keys ({len(js[0])} columns): {list(js[0].keys())}")
    except Exception as e:
        print(f"Error {name} paginated: {e}")

# 2. Test other categories (ETFs, BDRs, etc.)
for cat_type in [3, 4, 5, 6]:
    url = f"https://statusinvest.com.br/category/advancedsearchresultpaginated?CategoryType={cat_type}"
    payload = {'search': '{}', 'CategoryType': cat_type, 'page': 0, 'take': 1000}
    try:
        data = urllib.parse.urlencode(payload).encode('utf-8')
        req = urllib.request.Request(url, data=data, headers=headers)
        with urllib.request.urlopen(req, timeout=10) as resp:
            content = resp.read()
            js = json.loads(content)
            items = js.get('list', []) if isinstance(js, dict) else js
            print(f"Cat {cat_type}: {len(items)} items")
            if len(items) > 0:
                print(f"  First: {items[0].get('ticker')} - {items[0].get('companyname')}")
                print(f"  Keys: {list(items[0].keys())}")
    except Exception as e:
        print(f"Cat {cat_type}: {e}")
