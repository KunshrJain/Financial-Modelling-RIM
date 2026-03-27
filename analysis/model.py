import pandas as pd

def rim_valuation(row):
    bv = row['BookValue']
    profit = row['NetProfit']
    ke = row['CostOfEquity']
    
    equity_charge = bv * ke
    residual_income = profit - equity_charge
    intrinsic_value = bv + (residual_income / ke)
    return intrinsic_value

data = {
    'Symbol': ['INFY', 'TCS'],
    'BookValue': [85000, 102000],
    'NetProfit': [24000, 42000],
    'CostOfEquity': [0.13, 0.12]
}

df = pd.DataFrame(data)
df['IntrinsicValue'] = df.apply(rim_valuation, axis=1)

print(df[['Symbol', 'BookValue', 'IntrinsicValue']])