// ==UserScript==
// @name         AI Job Apply
// @namespace    http://tampermonkey.net/
// @version      2025-11-07
// @description  try to take over the world!
// @author       You
// @match        https://www.linkedin.com/jobs/search/*
// @match        https://www.linkedin.com/jobs/view/*
// @icon         https://www.google.com/s2/favicons?sz=64&domain=linkedin.com
// @grant        GM_xmlhttpRequest
// @grant        unsafeWindow
// @grant        GM_addElement
// ==/UserScript==

setTimeout(function(){
    (function() {
        'use strict';

        // --- CONFIGURATION ---
        // !! This MUST be updated with your ngrok URL !!
        const API_ROOT = "https://58d35aa75449.ngrok-free.app";
        let jobIndex = 0
        let jobs = extractJobs()
        setTimeout(openJob, 2000)
        // !! ----------------- !!

        function openJob(){


          let j = jobs[jobIndex]

          if(!j){
              console.log("set new page")
              clickNextPage()
              setTimeout(function(){
                 jobIndex = 0
                 jobs = extractJobs()
                 openJob()
              }, 4000)
              return
          }
          // 3. Find the clickable link element within that listing
          // The most reliable clickable element is the job title link.
          const jobLink = j.element.querySelector('.job-card-container__link');

            if (jobLink) {
                console.log(`Clicking the job link for: ${jobLink.textContent.trim()}`);
                setTimeout(function(){
                    //triggerAiApply()
                    clickEasyApplyButton()
                }, 3000)
                // 4. Use the .click() method to simulate a user click
                jobLink.click();
                jobIndex++
                return true;
            } else {
                console.error("The clickable link element was not found inside the first job listing.");
                return false;
            }
        }

         // 1. Create and inject the "Apply with AI" button
        const aiButton = document.createElement("button");
        aiButton.textContent = "Apply with AI";
        aiButton.style.position = "fixed";
        aiButton.style.bottom = "20px";
        aiButton.style.right = "20px";
        aiButton.style.zIndex = "109999";
        aiButton.style.backgroundColor = "#0a66c2";
        aiButton.style.color = "white";
        aiButton.style.padding = "10px 15px";
        aiButton.style.border = "none";
        aiButton.style.borderRadius = "8px";
        aiButton.style.cursor = "pointer";
        aiButton.style.fontSize = "16px";
        aiButton.style.boxShadow = "0 4px 10px rgba(0,0,0,0.2)";

        // Check if the DOM is ready before appending the button
        if (document.body) {
            document.body.appendChild(aiButton);
        } else {
            window.addEventListener('load', () => document.body.appendChild(aiButton));
        }


    // 2. Add click listener
    aiButton.addEventListener("click", triggerAiApply);

    /**
    * Looks for and clicks the "Easy Apply" button in the job details pane.
    */
    function clickEasyApplyButton() {
        'use strict';

        // Target the button using its specific classes and data attribute
        const easyApplyButton = document.querySelector(
            'button.jobs-apply-button[data-live-test-job-apply-button]'
        );

        if (easyApplyButton) {
            const buttonText = easyApplyButton.textContent.trim();
            console.log(`Found and clicking the: ${buttonText} button.`);
            easyApplyButton.click();
            setTimeout(function(){
                triggerAiApply()
            }, 5000)
            return true;
        } else {
            console.warn("The 'Easy Apply' button was not found in the current view.");
            return false;
        }
    }

    function extractJobs() {
         'use strict';

         // The selector targets the <li> element that contains the entire job card structure.
         const jobListings = document.querySelectorAll('li.occludable-update[data-occludable-job-id]');

         if (jobListings.length === 0) {
             console.warn("Could not find any job listings using the selector: li.occludable-update[data-occludable-job-id].");
             return;
         }

         console.log(`--- Found ${jobListings.length} Job Listings ---`);

         const extractedJobs = Array.from(jobListings).map((jobElement, index) => {
             // 1. Extract the unique job ID
             const jobId = jobElement.getAttribute('data-occludable-job-id');

             // 2. Extract the job title using a reliable nested selector
             const titleElement = jobElement.querySelector('.job-card-list__title--link');
             const title = titleElement ? titleElement.textContent.trim() : 'Title Not Found';

             // 3. Extract the company name
             const companyElement = jobElement.querySelector('.artdeco-entity-lockup__subtitle span');
             const company = companyElement ? companyElement.textContent.trim() : 'Company Not Found';

             console.log(`[${index + 1}] ID: ${jobId}, Title: ${title}, Company: ${company}`);

             return {
                 jobId: jobId,
                 title: title,
                 company: company,
                 element: jobElement // You can access the DOM element itself if needed
             };
         });

         console.log("--- Extraction Complete ---");
         // Returns the array of objects in case you want to use it programmatically
         return extractedJobs;

     }

     function clickNextPage() {
        // Target the button using its specific, stable aria-label
        const nextButton = document.querySelector('button[aria-label="View next page"]');

        if (nextButton) {
            console.log("Navigating to the next search results page...");
            nextButton.click();
            return true;
        } else {
            console.warn("Could not find the 'Next Page' button. Automation stopped (likely reached last page).");
            return false;
        }
    }


    // This function now uses GM_xmlhttpRequest and relies on callbacks
    function triggerAiApply() {
        console.log("AI Apply clicked!");
        setButtonState("loading", "Analyzing...");

        // 3. Get the HTML of the active application modal
        const jobModal = document.querySelector('[data-test-modal].jobs-easy-apply-modal');

        if (!jobModal) {
            /* setButtonState("error", "Modal not found!");
            console.error("Could not find the LinkedIn Easy Apply modal. Please ensure it is open.");
            // Reset button after a few seconds
            setTimeout(() => { setButtonState("idle", "Apply with AI"); }, 3000);*/
            setTimeout(closePostApplyModal, 4000)
            setTimeout(openJob, 7000)
            return;
        }

        // Send only the content of the modal, not the entire document
        const modalHTML = jobModal.innerHTML;

        // 4. Send the HTML to our Go backend using GM_xmlhttpRequest
        GM_xmlhttpRequest({
            method: 'POST',
            url: `${API_ROOT}/api/analyze`,
            headers: {
                'Content-Type': 'application/json',
                "ngrok-skip-browser-warning": "true"
            },
            data: JSON.stringify({ html: modalHTML }),

            onload: function(response) {
                try {
                    if (response.status !== 200) {
                        throw new Error(`Server error: Status ${response.status} - ${response.statusText}`);
                    }

                    const data = JSON.parse(response.responseText);

                    // 5. Receive and execute the JavaScript code
                    if (data.jsCode) {
                        console.log("Received JavaScript from AI. Executing...");
                        executeAiScript(data.jsCode);
                        setButtonState("success", "Fields Filled!");
                    } else {
                        throw new Error("No JavaScript code received from AI or empty response.");
                    }
                } catch (error) {
                    console.error("AI Apply Error (Response Handing):", error);
                    setButtonState("error", "Error!");
                }

                triggerAiApply()
                // Reset button after a few seconds
                setTimeout(() => { setButtonState("idle", "Apply with AI"); }, 3000);
            },

            onerror: function(response) {
                console.error("AI Apply Error (Network/CORS):", response);
                console.error("GM_xmlhttpRequest failed. Ensure Tampermonkey is running and the API is accessible via your ngrok URL.");
                setButtonState("error", "Error!");

                // Reset button after a few seconds
                setTimeout(() => { setButtonState("idle", "Apply with AI"); }, 3000);
            }
        });
    }

     /**
    * Finds the "Done" button on the "Application sent" modal and clicks it.
    * It waits for the modal to appear before proceeding.
    */
    function closePostApplyModal() {
        'use strict';

        // 1. Target the specific modal container after submission
        const modalSelector = '[aria-labelledby="post-apply-modal"]';
        const modal = document.querySelector(modalSelector);

        if (!modal) {
            console.warn("Post-apply modal container not found.");
            return false;
        }

        // 2. Find ALL buttons in the actionbar
        const actionbar = modal.querySelector('.artdeco-modal__actionbar');
        if (!actionbar) {
            console.warn("Post-apply modal actionbar not found.");
            return false;
    }

    let attempts = 0;
    const maxAttempts = 10; // Try for up to 2 seconds

    function waitForAndClickDone() {
            const buttons = actionbar.querySelectorAll('button');
            let doneButton = null;

            // Iterate through all buttons in the actionbar to find the one with the text "Done"
            for (const button of buttons) {
                if (button.textContent.trim() === 'Done') {
                    doneButton = button;
                    break;
                }
            }

            if (doneButton) {
                console.log("Application complete. Clicking 'Done' to close the modal.");
                doneButton.click();
                return true;
            }

            if (attempts < maxAttempts) {
                attempts++;
                setTimeout(waitForAndClickDone, 100); // Check every 100ms
            } else {
                console.warn("Post-apply 'Done' button not found after multiple attempts.");
            }
        }

        // Start the waiting process
        waitForAndClickDone();
    }

    function executeAiScript(jsCode) {
        // --- FINAL CRITICAL FIX FOR CSP VIOLATION ---
        // GM_addElement('script') is specifically designed by Tampermonkey to bypass
        // the most restrictive CSP rules (like nonces and hash checks) by injecting
        // the script in a way that the browser is forced to trust it.

        try {
            // The code must be wrapped in a self-executing function for security and scoping
            const wrappedCode = `(function() {
                try {
                    // Slight delay to ensure DOM is ready after previous user action
                    setTimeout(function() {
                        ${jsCode}
                    }, 50);
                } catch (e) {
                    console.error('AI-generated script execution error:', e);
                }
            })();`;

            GM_addElement('script', {
                textContent: wrappedCode,
                type: 'text/javascript'
            });

            console.log("AI script executed successfully via GM_addElement.");

        } catch (e) {
            console.error("Error executing AI-generated script via GM_addElement:", e);
            console.error("Automation Failure: Check console.");
        }
    }

    function setButtonState(state, text) {
        aiButton.textContent = text;
        aiButton.disabled = (state === "loading");

        switch (state) {
            case "loading":
                aiButton.style.backgroundColor = "#5a8ecf";
                aiButton.style.cursor = "wait";
                break;
            case "success":
                aiButton.style.backgroundColor = "#0f8a00";
                break;
            case "error":
                aiButton.style.backgroundColor = "#d90429";
                break;
            case "idle":
            default:
                aiButton.style.backgroundColor = "#0a66c2";
                aiButton.style.cursor = "pointer";
                break;
        }
    }



    })();
}, 5 *1000)