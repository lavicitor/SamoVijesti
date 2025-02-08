import "./style.css";

function fetchAndDisplayRSS(rssURL) {
  const cacheBuster = new Date().getTime();
  const urlWithCacheBuster = `${rssURL}?cb=${cacheBuster}`;
  window.go.main.App.GetRSS(urlWithCacheBuster)
    .then(function (items) {
      const itemList = document.getElementById("articleList");
      itemList.childNodes.forEach((child) => itemList.removeChild(child))
      itemList.innerHTML = "";

      if (items.length === 0) {
        itemList.innerHTML = "<p>No items found or invalid RSS feed.</p>";
        return;
      }

      let i = 0;
      function addItem() {
        if (i < items.length) {
          const item = items[i];
          const li = document.createElement("li");
          window.go.main.App.GetArticleImage(item.Link).then((ImageURL) => {
            item.ImageURL = ImageURL;
            li.style = "border-bottom: 1px dotted grey; list-style-type: none;";
            li.innerHTML = `<p><img src="${item.ImageURL}" width="100%"></p><h3><a href="${item.Link}" target="_blank">${item.Title}</a></h3><p>${item.Description}</p><p>Published: ${item.PubDate}</p>`;
            itemList.appendChild(li);
            i++;
            requestAnimationFrame(addItem);
          });
        }
      }
      addItem();
    })
    .catch(function (error) {
      console.error("Error fetching RSS:", error);
      alert("Error fetching RSS: " + error);
    });
}

const dvadeseticetirisataHrSubmenu = document.getElementById('dvadeseticetirisataHrSubmenu');
const indexHrSubmenu = document.getElementById('indexHrSubmenu');
const slobodnadalmacijaHrSubmenu = document.getElementById('slobodnadalmacijaHrSubmenu');

document.getElementById("dvadeseticetirisataHrButton").addEventListener("click", () => {
  indexHrSubmenu.style.display = 'none'
  slobodnadalmacijaHrSubmenu.style.display = 'none'
  dvadeseticetirisataHrSubmenu.style.display = dvadeseticetirisataHrSubmenu.style.display === 'block' ? 'none' : 'block';
});

document.getElementById("dvadeseticetirisataHrAktualno").addEventListener("click", () => {
  fetchAndDisplayRSS("http://www.24sata.hr/feeds/aktualno.xml");
});

document.getElementById("dvadeseticetirisataHrNajnovije").addEventListener("click", () => {
  fetchAndDisplayRSS("http://www.24sata.hr/feeds/najnovije.xml");
});

document.getElementById("dvadeseticetirisataHrNews").addEventListener("click", () => {
  fetchAndDisplayRSS("http://www.24sata.hr/feeds/news.xml");
});

document.getElementById("dvadeseticetirisataHrShow").addEventListener("click", () => {
  fetchAndDisplayRSS("http://www.24sata.hr/feeds/show.xml");
});

document.getElementById("dvadeseticetirisataHrSport").addEventListener("click", () => {
  fetchAndDisplayRSS("http://www.24sata.hr/feeds/sport.xml");
});

document.getElementById("dvadeseticetirisataHrLifestyle").addEventListener("click", () => {
  fetchAndDisplayRSS("http://www.24sata.hr/feeds/lifestyle.xml");
});

document.getElementById("dvadeseticetirisataHrTech").addEventListener("click", () => {
  fetchAndDisplayRSS("http://www.24sata.hr/feeds/tech.xml");
});

document.getElementById("dvadeseticetirisataHrViral").addEventListener("click", () => {
  fetchAndDisplayRSS("http://www.24sata.hr/feeds/fun.xml");
});

document.getElementById("indexHrButton").addEventListener("click", () => {
  dvadeseticetirisataHrSubmenu.style.display = 'none'
  slobodnadalmacijaHrSubmenu.style.display = 'none'
  indexHrSubmenu.style.display = indexHrSubmenu.style.display === 'block' ? 'none' : 'block';
});

document.getElementById("indexHrNajnovije").addEventListener("click", () => {
  fetchAndDisplayRSS("https://www.index.hr/rss");
});

document.getElementById("indexHrVijesti").addEventListener("click", () => {
  fetchAndDisplayRSS("https://www.index.hr/rss/vijesti");
});

document.getElementById("indexHrHrvatska").addEventListener("click", () => {
  fetchAndDisplayRSS("https://www.index.hr/rss/vijesti-hrvatska");
});

document.getElementById("indexHrZagreb").addEventListener("click", () => {
  fetchAndDisplayRSS("https://www.index.hr/rss/vijesti-zagreb");
});

document.getElementById("indexHrRegija").addEventListener("click", () => {
  fetchAndDisplayRSS("https://www.index.hr/rss/vijesti-regija");
});

document.getElementById("indexHrEU").addEventListener("click", () => {
  fetchAndDisplayRSS("https://www.index.hr/rss/vijesti-eu");
});

document.getElementById("indexHrSvijet").addEventListener("click", () => {
  fetchAndDisplayRSS("https://www.index.hr/rss/vijesti-svijet");
});

document.getElementById("indexHrZnanost").addEventListener("click", () => {
  fetchAndDisplayRSS("https://www.index.hr/rss/vijesti-znanost");
});

document.getElementById("indexHrCrnaKronika").addEventListener("click", () => {
  fetchAndDisplayRSS("https://www.index.hr/rss/vijesti-crna-kronika");
});

document.getElementById("indexHrNovac").addEventListener("click", () => {
  fetchAndDisplayRSS("https://www.index.hr/rss/vijesti-novac");
});

document.getElementById("slobodnadalmacijaHrButton").addEventListener("click", () => {
  dvadeseticetirisataHrSubmenu.style.display = 'none'
  indexHrSubmenu.style.display = 'none'
  slobodnadalmacijaHrSubmenu.style.display = slobodnadalmacijaHrSubmenu.style.display === 'block' ? 'none' : 'block';
});

document.getElementById("slobodnadalmacijaHrNajnovije").addEventListener("click", () => {
  fetchAndDisplayRSS("https://slobodnadalmacija.hr/feed");
});

document.getElementById("slobodnadalmacijaHrVijesti").addEventListener("click", () => {
  fetchAndDisplayRSS("https://slobodnadalmacija.hr/feed/category/119");
});

document.getElementById("slobodnadalmacijaHrHrvatska").addEventListener("click", () => {
  fetchAndDisplayRSS("https://slobodnadalmacija.hr/feed/category/142");
});

document.getElementById("slobodnadalmacijaHrSvijet").addEventListener("click", () => {
  fetchAndDisplayRSS("https://slobodnadalmacija.hr/feed/category/241");
});

document.getElementById("slobodnadalmacijaHrPolitika").addEventListener("click", () => {
  fetchAndDisplayRSS("https://slobodnadalmacija.hr/feed/category/242");
});

document.getElementById("slobodnadalmacijaHrCrnaKronika").addEventListener("click", () => {
  fetchAndDisplayRSS("https://slobodnadalmacija.hr/feed/category/243");
});

document.getElementById("slobodnadalmacijaHrBiznis").addEventListener("click", () => {
  fetchAndDisplayRSS("https://slobodnadalmacija.hr/feed/category/244");
});

document.getElementById("slobodnadalmacijaHrRegija").addEventListener("click", () => {
  fetchAndDisplayRSS("https://slobodnadalmacija.hr/feed/category/245");
});

document.getElementById("telegramHrButton").addEventListener("click", () => {
  dvadeseticetirisataHrSubmenu.style.display = 'none'
  indexHrSubmenu.style.display = 'none'
  slobodnadalmacijaHrSubmenu.style.display = 'none'
  fetchAndDisplayRSS("https://www.telegram.hr/feed/");
});

document.addEventListener("DOMContentLoaded", () => {
  const articleList = document.getElementById("articleList");
  const modal = document.getElementById("modal");
  const modalContent = document.getElementById("modal-body");
  const closeBtn = document.querySelector(".close");

  articleList.addEventListener("click", (event) => {
    const link = event.target.closest("a");
    if (link) {
      event.preventDefault(); // Prevent default link behavior

      const url = link.href;
      console.log(url);

      // Open the modal
      modal.style.display = "block";
      document.body.style.overflow = "hidden";

      window.go.main.App.GetArticleContent(url)
        .then((content) => {
          modalContent.innerHTML = content;
        })
        .catch((error) => {
          console.error("Error getting article content:", error);
          modalContent.innerHTML = "<p>Error loading content.</p>";
        });
    }
  });

  // Close the modal
  closeBtn.addEventListener("click", () => {
    modal.style.display = "none";
    modalContent.innerHTML = ""; // Clear modal content when closed.
    document.body.style.overflow = "auto";
  });

  // Close the modal if the user clicks outside of the modal content
  window.addEventListener("click", (event) => {
    if (event.target == modal) {
      modal.style.display = "none";
      modalContent.innerHTML = ""; // Clear modal content when closed.
      document.body.style.overflow = "auto";
    }
  });
});
