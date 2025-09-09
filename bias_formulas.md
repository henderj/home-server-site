# Dice Bias Quantification Formulas

This document outlines the statistical formulas required to quantify bias in a die based on observed roll frequencies. Each formula is presented with its mathematical expression and a description of its purpose in assessing whether a die (default six-sided, with faces 1 to \( k \), where \( k = 6 \) unless specified) deviates from a uniform distribution (fairness). These formulas are designed for a feature that processes an array of roll outcomes and produces metrics for overall bias and bias toward specific faces.

---

## 1. Data Summarization Formulas

### Total Number of Rolls (\( n \))
- **Formula**: \( n = \) length of the input array of roll outcomes.
- **Purpose**: Determines the sample size, which is the total count of dice rolls observed. This is the basis for frequency calculations.

### Observed Frequency (\( O_i \))
- **Formula**: \( O_i = \) count of rolls equal to face \( i \), for \( i = 1, 2, \ldots, k \).
- **Purpose**: Quantifies how many times each face (e.g., 1, 2, ..., 6 for a six-sided die) appeared in the data. These are the observed counts compared against expected counts.

### Expected Frequency (\( E_i \))
- **Formula**: \( E_i = \frac{n}{k} \), for each face \( i = 1, 2, \ldots, k \).
- **Purpose**: Represents the expected number of rolls for each face under the null hypothesis of a fair die, where each face has equal probability \( \frac{1}{k} \). Used as the baseline for fairness.

---

## 2. Chi-Squared Goodness-of-Fit Test

### Chi-Squared Statistic (\( \chi^2 \))
- **Formula**: 
  \[
  \chi^2 = \sum_{i=1}^{k} \frac{(O_i - E_i)^2}{E_i}
  \]
- **Purpose**: Measures the overall deviation of observed frequencies (\( O_i \)) from expected frequencies (\( E_i \)) for a fair die. A larger value indicates greater evidence of bias (non-uniformity).

### Degrees of Freedom (\( df \))
- **Formula**: \( df = k - 1 \).
- **Purpose**: Specifies the degrees of freedom for the chi-squared distribution, based on the number of faces (\( k \)). Used to interpret the chi-squared statistic.

### P-Value
- **Formula**: \( p = 1 - \text{CDF}(\chi^2, df) \), where CDF is the cumulative distribution function of the chi-squared distribution with \( df \) degrees of freedom.
- **Purpose**: Quantifies the probability of observing a chi-squared statistic at least as extreme as the computed \( \chi^2 \) under the null hypothesis (die is fair). A p-value < 0.05 indicates significant evidence of bias (reject fairness). Requires a statistical library or gamma function approximation for calculation.

### Warning for Small Samples
- **Condition**: If \( n < 30 \) or any \( E_i < 5 \).
- **Purpose**: Flags unreliable results due to small sample size or low expected frequencies, where chi-squared test approximations may fail. Suggests collecting more data.

---

## 3. Overall Bias Quantification

### Cramér's V (\( V \))
- **Formula**: 
  \[
  V = \sqrt{\frac{\chi^2}{n \times (k - 1)}}
  \]
- **Purpose**: Provides a normalized effect size for the degree of bias, ranging from 0 (no bias) to 1 (maximum bias). Interpret as: <0.1 (negligible), 0.1-0.3 (small), 0.3-0.5 (medium), >0.5 (large). Accounts for sample size and number of faces.

### Estimated Probabilities (\( p_i \))
- **Formula**: \( p_i = \frac{O_i}{n} \), for each face \( i = 1, 2, \ldots, k \).
- **Purpose**: Estimates the empirical probability of rolling each face based on observed data. Used to compare against the fair probability \( \frac{1}{k} \).

### Total Variation Distance (\( TVD \))
- **Formula**: 
  \[
  TVD = \frac{1}{2} \sum_{i=1}^{k} \left| p_i - \frac{1}{k} \right|
  \]
- **Purpose**: Measures the overall bias as the total deviation of estimated probabilities (\( p_i \)) from the fair probability (\( \frac{1}{k} \)). Ranges from 0 (perfectly fair) to 1 (maximum bias). Provides an intuitive [0,1] bias score.

---

## 4. Bias Toward Specific Numbers

### Standardized Residuals (\( r_i \))
- **Formula**: 
  \[
  r_i = \frac{O_i - E_i}{\sqrt{E_i}}, \text{ for each face } i
  \]
- **Purpose**: Identifies which specific faces contribute to bias. Positive \( r_i > 2 \) indicates bias
