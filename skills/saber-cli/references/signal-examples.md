# Signal Examples by Industry and Use Case

Example signal questions organized by industry and buyer persona. Use these as
starting points when helping users define their buying signals during discovery.

## SaaS / Software

**Company signals:**
- "Is this company actively hiring in sales or revenue roles?"
- "Has this company recently raised a funding round?"
- "Is this company migrating or evaluating new CRM software?"
- "Does this company use Salesforce as their CRM?"
- "Is this company expanding into new markets or geographies?"
- "Has this company recently appointed a new CTO or VP Engineering?"

**Contact signals:**
- "Is this person posting about scaling their sales team?"
- "Has this person recently changed roles or been promoted?"
- "Is this person engaging with content about sales automation?"

## HR / People Tech

**Company signals:**
- "Is this company actively hiring in HR or people operations?"
- "Does this company have open roles for HRIS administrators?"
- "Is this company experiencing high employee turnover?"
- "Has this company been mentioned in news about workplace culture initiatives?"

**Contact signals:**
- "Is the Head of People posting about employee retention challenges?"
- "Is this person sharing content about HR technology modernization?"
- "Has this HR leader recently spoken at industry events?"

## Cybersecurity

**Company signals:**
- "Has this company experienced a data breach or security incident recently?"
- "Is this company hiring for security engineering or CISO roles?"
- "Does this company operate in a regulated industry requiring compliance certifications?"
- "Is this company evaluating or migrating their cloud security posture?"

**Contact signals:**
- "Is this CISO posting about zero-trust architecture?"
- "Has this security leader commented on recent industry vulnerabilities?"

## Financial Services / Fintech

**Company signals:**
- "Is this company launching new digital banking or payment products?"
- "Is this company hiring for compliance or regulatory roles?"
- "Has this company received any regulatory actions or fines recently?"
- "Is this company expanding their API or platform partnerships?"

**Contact signals:**
- "Is this person posting about open banking or embedded finance?"
- "Has this product leader shared content about payment infrastructure?"

## E-commerce / Retail

**Company signals:**
- "Is this company expanding their online presence or launching new stores?"
- "Is this company hiring for e-commerce or digital marketing roles?"
- "Has this company recently changed their e-commerce platform?"
- "Is this company running large promotional campaigns?"

**Contact signals:**
- "Is this person posting about omnichannel retail strategies?"
- "Has this marketing leader shared content about customer acquisition costs?"

## Healthcare / Life Sciences

**Company signals:**
- "Is this company hiring for clinical or research roles?"
- "Has this company received FDA approval or regulatory clearance recently?"
- "Is this company expanding into telehealth or digital health?"
- "Is this company involved in clinical trials for new treatments?"

**Contact signals:**
- "Is this person publishing about healthcare data interoperability?"
- "Has this medical director posted about patient care innovations?"

## Manufacturing / Industrial

**Company signals:**
- "Is this company investing in automation or Industry 4.0 technologies?"
- "Is this company expanding manufacturing capacity or opening new facilities?"
- "Has this company been affected by supply chain disruptions?"
- "Is this company hiring for operations or supply chain roles?"

**Contact signals:**
- "Is this operations leader posting about lean manufacturing or efficiency?"
- "Has this person shared content about sustainability in manufacturing?"

## Signal Design Tips

When crafting signal questions:

1. **Be specific** -- "Is this company actively hiring SDRs?" is better than "Is this company growing?"
2. **Target observable behavior** -- signals should be answerable from public information (job postings, news, social posts, filings)
3. **Link to buying intent** -- every signal should connect to why this company might need your product
4. **Use boolean for prioritization** -- `boolean` answer type makes it easy to sort and filter
5. **Use open_text for discovery** -- `open_text` gives richer answers when exploring a new market
6. **Test before scaling** -- run a sample of 5-10 companies first to validate signal quality

## Answer Type Selection Guide

| Use case | Answer type | Example |
|---|---|---|
| Yes/no qualification | `boolean` | "Are they hiring engineers?" |
| Free-form research | `open_text` | "What is their tech stack?" |
| Quantitative comparison | `number` | "How many open roles do they have?" |
| Multiple items | `list` | "What products do they sell?" |
| Growth metrics | `percentage` | "What is their YoY revenue growth?" |
| Budget/revenue sizing | `currency` | "What is their estimated ARR?" |
| Find a specific page | `url` | "Where is their careers page?" |
