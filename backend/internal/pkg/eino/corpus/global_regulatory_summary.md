# Global Food & Candy Regulatory Summary for AI / RAG Context

This document serves as a quick reference for global food safety, labeling, and import regulations, specifically tailored for confectionery and candy exports.

## North America & South America
| Abbreviation / Agency | Role | Official / Authority Website |
| :--- | :--- | :--- |
| **FDA / FSMA** (USA) | US Food and Drug Administration. Oversees the Food Safety Modernization Act (FSMA) (2011), the core food safety legal framework. Manages most foods, including candy. | [FDA](https://www.fda.gov) (FSMA: "Food" -> "Food Safety Modernization Act") |
| **CFIA** (Canada) | Canadian Food Inspection Agency. Responsible for food safety and labeling, notably mandatory English/French bilingual labeling. | [CFIA](https://inspection.canada.ca) |
| **ANVISA** (Brazil) | National Health Surveillance Agency. Responsible for food (including novel foods) registration and supervision. | [ANVISA](https://www.gov.br/anvisa/pt-br) |
| **ANMAT** (Argentina) | National Administration of Drugs, Foods and Medical Devices. Key entity for food labeling and safety. | [ANMAT](https://www.argentina.gob.ar/anmat) |

## Major European Food Safety Standards
| Standard | Description | Official / Authority Website |
| :--- | :--- | :--- |
| **IFS** | International Featured Standards. Driven by German/French retailers, GFSI recognized. Widely used for EU food factory audits, including candy. | [IFS](https://www.ifs-certification.com) |
| **BRC / BRCGS** | BRCGS Global Standard for Food Safety. Originated from British Retail Consortium, GFSI recognized. Adopted by numerous European/UK processed food factories. | [BRCGS](https://www.brcgs.com) |
| **EU Food Labeling** | e.g., Regulation (EU) No 1169/2011. Specifies nutrition labeling, allergens, ingredient lists. | [EUR-Lex](https://eur-lex.europa.eu) (Search "1169/2011") |

## Middle East & Halal Requirements
| Country / Region | Agency / Key Points | Official / Authority Website |
| :--- | :--- | :--- |
| **Saudi Arabia** | **SFDA + Saudi Halal Center**: Responsible for food safety and imported Halal certification (critical for gelatin-containing chocolates and candies). | [SFDA](https://sfda.gov.sa); [Halal Center](https://halal.gov.sa) |
| **Other Gulf Countries** (UAE, Kuwait, Qatar, etc.) | National standards agencies + Halal systems. UAE previously led by ESMA, now integrated into a broader regulatory system. Halal is mandatory. | [ESMA (UAE)](https://www.esma.gov.ae) |

## Asia-Pacific Key Agencies
| Country / Region | Agency / Standard | Official / Authority Website |
| :--- | :--- | :--- |
| **Japan** | **JAS** (Japanese Agricultural Standards). Managed by MAFF (Ministry of Agriculture, Forestry and Fisheries). Includes organic JAS. | [MAFF](https://www.maff.go.jp/e/policies/standard/) |
| **China** | **GACC** (General Administration of Customs). Responsible for imported food registration, inspection, quarantine, and customs clearance. | [GACC](http://www.customs.gov.cn) |
| **South Korea** | **MFDS** (Ministry of Food and Drug Safety). Responsible for food and additive standards. | [MFDS](https://www.mfds.go.kr/eng) |
| **India** | **FSSAI** (Food Safety and Standards Authority of India). Unified food safety and labeling standards. | [FSSAI](https://www.fssai.gov.in) |
| **Indonesia** | **BPOM** (National Agency of Drug and Food Control) + Halal System (**BPJPH / LPPOM MUI**). Halal certification is mandatory. | [BPOM](https://www.pom.go.id); BPJPH under Ministry of Religious Affairs. |
| **Singapore** | **SFA** (Singapore Food Agency). Food safety and supply authority. | [SFA](https://www.sfa.gov.sg) |
| **Hong Kong** | **CFS** (Centre for Food Safety). Manages prepackaged food labeling and safety. | [CFS](https://www.cfs.gov.hk) |
| **Australia & NZ** | **FSANZ** (Food Standards Australia New Zealand). Develops Food Standards Code, including sweetener regulations. | [FSANZ](https://www.foodstandards.gov.au/) |

## Africa & Other Regions
| Country / Region | Agency | Official / Authority Website |
| :--- | :--- | :--- |
| **Nigeria** | **NAFDAC**. Regulates food and drugs. | [NAFDAC](https://www.nafdac.gov.ng) |
| **South Africa** | **SABS** (South African Bureau of Standards). National standards body. | [SABS](https://www.sabs.co.za) |
| **Kenya** | **KEBS** (Kenya Bureau of Standards). National standards and certification. | [KEBS](https://www.kebs.org) |
| **Morocco** | **ONSSA** (National Office of Food Safety). National food safety agency. | [ONSSA](http://www.onssa.gov.ma) |

---
**Configuration Recommendations for AI Integration:**
*   **USA**: `{ regulator: "FDA", has_fsma: true, kosher_optional: true }`
*   **Saudi Arabia**: `{ regulator: "SFDA", halal_mandatory: true, arabic_label_required: true }`
*   **Indonesia**: `{ regulator: "BPOM", halal_authority: "BPJPH/MUI", halal_mandatory: true }`
*   **Malaysia**: `{ regulator: "MOH", halal_authority: "JAKIM", halal_mandatory: true }`
*   **Canada**: `{ regulator: "CFIA", bilingual_label_required: true }`
*   **Russia**: `{ regulator: "EAC", custom_notes: "Consider sanctions, payment, and logistics restrictions" }`
*   **Europe**: `{ certificates_preferred: ["IFS", "BRCGS"] }`
