import urllib.request
import json

url = "http://localhost:8080/api/v1/rankings"
req = urllib.request.urlopen(url)
data = json.loads(req.read().decode('utf-8'))

acoes = data.get('acoes', [])
fiis = data.get('fiis', [])

print(f"Total Acoes: {len(acoes)} | Total FIIs: {len(fiis)}")
print("\n" + "="*80)
print("--- ACOES COM DY > 15% (POSSIVEIS YIELD TRAPS / ATIPICOS) ---")
print("="*80)

high_dy_acoes = [a for a in acoes if a.get('dy', 0) > 15]
high_dy_acoes.sort(key=lambda x: x.get('dy', 0), reverse=True)

for a in high_dy_acoes[:25]:
    ticker = a.get('ticker')
    nome = a.get('nome', '')[:25]
    preco = a.get('preco_atual', 0)
    dy = a.get('dy', 0)
    bazin = a.get('preco_teto_bazin', 0)
    graham = a.get('valor_graham', 0)
    score = a.get('score', 0)
    vol = a.get('volume_total', 0)
    pl = a.get('pl', 0)
    lpa = a.get('lpa', 0)
    div12m = a.get('dividendos_12m', 0)
    payout = (div12m / lpa * 100) if lpa > 0 else 0
    print(f"{ticker:<6} | {nome:<25} | Preco: R${preco:>6.2f} | DY: {dy:>5.1f}% | Div12M: R${div12m:>5.2f} | LPA: R${lpa:>5.2f} | Payout: {payout:>5.1f}% | Score: {score:>3} | Vol: R${vol:>10,.0f}")

print("\n" + "="*80)
print("--- FIIS COM DY > 15% ---")
print("="*80)
high_dy_fiis = [f for f in fiis if f.get('dy', 0) > 15]
high_dy_fiis.sort(key=lambda x: x.get('dy', 0), reverse=True)
for f in high_dy_fiis[:15]:
    ticker = f.get('ticker')
    nome = f.get('nome', '')[:25]
    preco = f.get('preco_atual', 0)
    dy = f.get('dy', 0)
    pvp = f.get('pvp', 0)
    score = f.get('score', 0)
    vol = f.get('volume_total', 0)
    print(f"{ticker:<6} | {nome:<25} | Preco: R${preco:>6.2f} | DY: {dy:>5.1f}% | P/VP: {pvp:>4.2f} | Score: {score:>3} | Vol: R${vol:>10,.0f}")
