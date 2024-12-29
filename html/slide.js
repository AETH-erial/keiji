/* Set the width of the side navigation to 250px and the left margin of the page content to 250px */
function toggleNav() {
  if (document.getElementById("mySidenav").style.width === "400px"){
      document.getElementById("mySidenav").style.width = "0px";
      document.getElementById("main").style.marginLeft = "0px";
      return;
  } else {
      document.getElementById("mySidenav").style.width = "400px";
      document.getElementById("main").style.marginLeft = "400px";
      return;
  }
}

