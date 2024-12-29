sessionStorage.clear();

document.addEventListener("DOMContentLoaded", () => {
  const form = document.getElementById("login-form");
  const responseMessage = document.getElementById("response-message");

  form.addEventListener("submit", async function (event) {
    event.preventDefault();

    const requestBody = {
      u: document.getElementById("user").value,
      p: document.getElementById("pass").value,
      csrf: document
        .querySelector('meta[name="csrf-token"]')
        .getAttribute("content"),
    };

    try {
      const response = await fetch("/api/v1/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(requestBody),
      });

      if (!response.ok) {
        const error = await response.json();
        responseMessage.style.display = "block";
        responseMessage.textContent = error.message || "Login failed";
        return;
      }

      const responseData = await response.json();
      if (responseData.userData) {
        sessionStorage.setItem(
          "userData",
          JSON.stringify(responseData.userData)
        );
      }

      window.location.href = "/dashboard";
    } catch (error) {
      responseMessage.style.display = "block";
      responseMessage.textContent = "An error occurred: " + error.message;
    }
  });
});
