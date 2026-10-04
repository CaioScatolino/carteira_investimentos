import json, urllib.request

req = urllib.request.urlopen('http://localhost:8080/api/v1/rankings')
data = json.loads(req.read().decode('utf-8'))
for ticker in ['RIAA3', 'GRND3', 'HBRE3', 'SCAR3', 'LOGG3', 'PATL11', 'BBFI11', 'HGPO11']:
    item = next((x for x in data['acoes'] + data['fiis'] if x['ticker'] == ticker), None)
    if item:
        print(f"{item['ticker']:6} | Preco: R${item['preco_atual']:6.2f} | DY: {item['dy']:6.1f}% | Score Atual: {item['score']} | Status: {item['status']}")
