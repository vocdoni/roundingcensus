## Summary

**Config**
| Param | Value |
|:---|---:|
| `MIN_PRIVACY_THRESHOLD` | 3 |
| `MIN_ACCURACY` (%) | 50 |
| `GROUP_BALANCE_DIFF` | 10 |

**Command**

```sh
TEST_CENSUS=./holders.json MIN_PRIVACY_THRESHOLD=3 MIN_ACCURACY=50 GROUP_BALANCE_DIFF=10  go test -timeout 30s -v -run ^TestAutoRoundingAlgorithm$
```

**Round results**

| Token name | Type | Holders | Accuracy | Groups | Time |
|:---|:---|:---:|:---:|:---:|:---:|
| API3 | ERC20 | 22292 | 54.28% | 2866 | 1.72s |
| CABIN | ERC20 | 723 | 96.66% | 116 | 0.04s |
| CREAM | ERC20 | 9079 | 31.00% | 1500 | 0.53s |
| FLEX | ERC20 | 82 | 72.59% | 13 | 0.000s |
| FWB | ERC20 | 7343 | 86.43% | 658 | 0.48s |
| MOONBIRDS | ERC721 | 5883 | 69.85% | 4 | 0.20s |
| OM | ERC20 | 7609 | 90.8% | 1738 | 0.54s |
| POINTS | ERC20 | 2150 | 97.29% | 143 | 0.13s |
| POAP:onvote-global-census-2023 | ERC721 | 119 | 100% | 1 | 0.00s |
| PRAY | ERC20 | 1401 | 96.99% | 13 | 0.08s |
| SURGE | ERC721 | 2581 | 58.24% | 3 | 0.09s |
| UNI | ERC20 | 115812 | 78.98% | 15227 | 14.00s |
| YAM | ERC20 | 11940 | 53.77% | 1177 | 0.74s |
