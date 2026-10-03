import urllib.request
import re

headers = {
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36',
    'Referer': 'https://statusinvest.com.br/acoes/busca-avancada'
}

req = urllib.request.Request('https://statusinvest.com.br/js/pages/home.advancedSearch.min.js?v=2.4.31.ABCDEG', headers=headers)
js_code = urllib.request.urlopen(req).read().decode('utf-8')
print('JS length:', len(js_code))

for m in re.finditer(r'advancedsearchresultpaginated[\s\S]{0,400}', js_code):
    print("Match paginated:", m.group(0))

for m in re.finditer(r'AdvancedSearchResultExport[\s\S]{0,400}', js_code):
    print("Match export:", m.group(0))

for m in re.finditer(r'CategoryType[\s\S]{0,200}', js_code):
    print("Match CategoryType:", m.group(0))
