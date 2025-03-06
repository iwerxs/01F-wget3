## Notes for WGET3

### Steps to Integrate the Rate-Limit Code
-  Define the -rate-limit Flag in main()
-  Add the flag parsing logic where you're defining other flags.
-  Call all parseRateLimit() to validate and convert the rate-limit value.
-  Pass the rateLimiter to the downloadFile() function to enforce speed limits.
-  Modify the downloadFile() function to apply the rate limiter using rateLimitedReader.
