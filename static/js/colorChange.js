/* Made by Tailwind Toolbox for the Landing Page template:
https://github.com/tailwindtoolbox/Landing-Page */

const header = document.getElementById("header");
const navcontent = document.getElementById("nav-content");
const navaction = document.getElementById("navAction");
const toToggle = document.querySelectorAll(".toggleColour");
const logo = document.getElementById("robologo");
let prevScrollPos = -1;

// Apply classes for slide in bar
document.addEventListener("scroll", () => {
	const scrollpos = window.scrollY;

	// Only make changes if scroll position has changed enough
	if (
		(scrollpos > 10 && prevScrollPos > 10) ||
		(scrollpos <= 10 && prevScrollPos <= 10)
	) {
		prevScrollPos = scrollpos;
		return;
	}

	if (scrollpos > 10) {
		logo.setAttribute("src", "static/images/robot-outline-black.svg");
		header.classList.add("bg-white");
		navaction.classList.remove("bg-white");
		navaction.classList.add("gradient");
		navaction.classList.remove("text-gray-800");
		navaction.classList.add("text-white");
		//Use to switch toggleColour colours
		for (let i = 0; i < toToggle.length; i++) {
			toToggle[i].classList.add("text-gray-800");
			toToggle[i].classList.remove("text-white");
		}
		header.classList.add("shadow");
		navcontent.classList.remove("bg-gray-100");
		navcontent.classList.add("bg-white");
	} else {
		logo.setAttribute("src", "static/images/robot-outline-white.svg");
		header.classList.remove("bg-white");
		navaction.classList.remove("gradient");
		navaction.classList.add("bg-white");
		navaction.classList.remove("text-white");
		navaction.classList.add("text-gray-800");
		//Use to switch toggleColour colours
		for (let i = 0; i < toToggle.length; i++) {
			toToggle[i].classList.add("text-white");
			toToggle[i].classList.remove("text-gray-800");
		}
		header.classList.remove("shadow");
		navcontent.classList.remove("bg-white");
		navcontent.classList.add("bg-gray-100");
	}
	prevScrollPos = scrollpos;
});
