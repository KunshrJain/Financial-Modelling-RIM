import os
import re
import pandas as pd
import matplotlib.pyplot as plt

# 1. Macro Settings (India March 2026)
RF_RATE = 0.071  # 10Y Govt Bond Yield
MARKET_RETURN = 0.13  # Expected Nifty Return
BETA_MAP = {'INFY': 1.10, 'TCS': 0.90, 'RELIANCE': 1.05}

def extract_from_xml(filepath):
    """Interrogates the raw XML for specific Ind-AS tags."""
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Regex to find numbers inside specific financial tags
    def get_tag_value(tag):
        pattern = rf'(?i){tag}[^>]*>([\d\s\.,]+)</'
        match = re.search(pattern, content)
        if match:
            # Clean commas and spaces from the raw XML string
            clean_val = match.group(1).replace(',', '').strip()
            return float(clean_val)
        return 0.0

    profit = get_tag_value("ProfitLossForPeriod")
    equity = get_tag_value("EquityAttributableToOwnersOfParent")
    return equity, profit

def run_evaluator():
    # Points to the folder one step above 'analysis/'
    data_dir = os.path.join(os.getcwd(), '..', 'data')
    results = []

    print(f"📂 Scanning for data in: {data_dir}")

    for file in os.listdir(data_dir):
        if file.endswith(".xml"):
            symbol = file.split('_')[0]
            equity, profit = extract_from_xml(os.path.join(data_dir, file))
            
            if equity > 0:
                beta = BETA_MAP.get(symbol, 1.0)
                # RIM Math
                ke = RF_RATE + beta * (MARKET_RETURN - RF_RATE)
                equity_charge = (equity / 1e7) * ke
                res_income = (profit / 1e7) - equity_charge
                # Intrinsic Value (Perpetuity)
                intrinsic_val = (equity / 1e7) + (res_income / ke)
                
                results.append({
                    'Symbol': symbol,
                    'ResidualIncome': res_income,
                    'IntrinsicValue': intrinsic_val,
                    'CapitalEfficiency': (profit / equity) * 100
                })

    df = pd.DataFrame(results)
    pd.options.display.float_format = '{:,.2f}'.format
    print("\n--- Phase 2: Residual Income Evaluation ---")
    print(df)

    # 2. Visualizing Efficiency vs Valuation
    plt.figure(figsize=(10, 6))
    plt.bar(df['Symbol'], df['IntrinsicValue'], color='teal', alpha=0.7, label='Intrinsic Value (Cr)')
    plt.title("Intrinsic Value: Absolute Valuation based on XBRL Data")
    plt.ylabel("Value in Crores (₹)")
    plt.savefig('valuation_summary.png')
    print("\n🚀 Plot saved as valuation_summary.png")

if __name__ == "__main__":
    run_evaluator()