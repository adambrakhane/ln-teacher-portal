const TOKEN_KEY = "token"


let main = {
    scheduleSelector: $("table#schedule"),
    scheduleTable: null,
    verifyUser: function() {

        // If we have a token, we're logged in
        // We could also check if the token has expired
        if (localStorage.getItem(TOKEN_KEY) != null) {
            console.log("user logged in")
            return
        } else {
            console.log("no token in storage")
            window.location = "/login.html"
        }
    },

    logOut: function() {
        localStorage.removeItem(TOKEN_KEY)
        window.location = "/login.html"
    },

    loadTeacherData: function(teacherID) {
        teacherID = parseInt(teacherID)

        // Attempt login
        jQuery.ajax({
                url: "/teachers/" + teacherID,
                type: "GET",
                headers: {
                    "Authorization": localStorage.getItem(TOKEN_KEY),
                },
            })
            .done(function(data, textStatus, jqXHR) {
                // 200 response, so we're in!
                console.log("HTTP Request Succeeded: " + jqXHR.status);
                console.log(data);
                data = JSON.parse(data)

                // Set name
                $("#teacherName").text(data.FullName)

                // Fill schedule
                $.each(data.Classes, function() {
                    var className = this.Name
                        // console.log("Class: ")
                        // console.log(this)
                    $.each(this.Schedule, function() {
                        // console.log(this)

                        this.className = className

                        main.scheduleTable.row.add(this).draw()
                    })
                })

            })
            .fail(function(jqXHR, textStatus, errorThrown) {
                console.log("HTTP Request Failed");
                console.log(textStatus)
            })
            .always(function() {
                /* could show loading bar here or something */
            });
    },
    cardAlert: function(config) { // type, header, body, ttl
        if (typeof config.type == "undefined") {
            console.log("must define alert type")
        }
        type = config.type
        switch (type) {
            case "alert":
            case "warning":
                type = "warning"
                break
            case "error":
            case "danger":
                type = "danger"
                break
            case "success":
                type = "success"
                break
            default:
                type = "dark"
        }

        head = ""
        seperator = "<hr>"
        body = ""

        if (typeof config.body == "undefined") {
            seperator = ""
            body = ""
        } else {
            body = config.body
        }
        if (typeof config.header == "undefined") {
            head = type
        } else {
            head = "<strong>" + config.header + "</strong>"
        }

        alert = $(`
        <div class="alert alert-` + type + ` alert-dismissible show cardError" role="alert">
            ` + head + `
            ` + seperator + `
            ` + body + `
        </div>        
        `)
        $("main").prepend(alert)

        // Hide it soon
        if (typeof config.ttl == "number") {
            var hideIt = function(alert, ttl) { // Do this to copy by value the alert box
                setTimeout(function() {
                    alert.slideUp()
                }, ttl * 1000)
            }
            hideIt(alert, config.ttl)
        }

    },
}

// Run on page load
$(document).ready(function() {

    // Make sure we're logged in
    main.verifyUser()

    // Page modules
    initScheduleTable()

    // Load teacher's ID from token
    var tokenData = parseJWT(localStorage.getItem(TOKEN_KEY))
    var teacherID = tokenData.teacher_id

    // Get teachers data and fill screen
    main.loadTeacherData(teacherID)

    // Attach button handlers
    $(".signOut").click(function(e) {
        e.preventDefault()
        main.logOut()
        return false
    })

    // Survey dialog
    $("table#schedule").on('click', 'button.survey', function(e) {
        console.log("survey")
        var dialog = $("#workshopFeedbackForm")

        // Get workshop data
        row = main.scheduleTable.row($(this).parents('tr'))
        var data = row.data();

        // Display the dialog
        dialog.find("#workshopFeedbackFormName").text(data.Workshop.Name)
        dialog.dialog()
    })
    $("#workshopFeedbackForm").on('click', 'button.submitSurvey', function(e) {
        console.log("survey close ")
        var dialog = $("#workshopFeedbackForm")
        dialog.dialog("destroy")
        main.cardAlert({ type: "success", header: "Thank you", body: "Your feedback is appreciated", ttl: 5 })
    })


})


function parseJWT(token) {
    console.log("Parsing token: " + token)
    var base64Url = token.split('.')[1];
    var base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    var jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
        return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
    }).join(''));

    return JSON.parse(jsonPayload);
};


function initScheduleTable() {

    main.scheduleTable = main.scheduleSelector.DataTable({
        data: [],
        lengthMenu: [
            [10, 25, 50, -1],
            [10, 25, 50, "all"]
        ],
        order: [
            [1, "asc"],
        ],
        pageLength: 25,
        select: {
            style: 'single'
        },
        rowId: "ID",
        columns: [
            { title: "Class", data: "className" },
            { title: "Name", data: "Workshop.Name" },
            {
                title: "Scheduled Date",
                data: "ScheduledAt",
                render: function(x) {
                    // return x
                    // Display dates nicely
                    return x.toLocaleString("en-US")
                }
            },
            // { title: "Name", data: "Name" },
            { title: "Description", data: "Workshop.Description" },
            {
                title: "Resources",
                data: "Resources",
                render: function(x) {
                    var out = "<ul>"

                    $.each(x, function() {
                        out += "<li><b>" + this.Name + ":</b> " + this.Description + "</li>"
                    })

                    out += "</ul>"

                    return out
                }
            },
            {
                title: "",
                searchable: false,
                sortable: false,
                render(data, type, full, meta) {
                    // If in the past, then show review link
                    if (Date.parse(full.ScheduledAt) < new Date()) {
                        // And add the button html
                        return `
                        <button class="btn btn-warning btn-sm survey">
                            Give feedback on this workshop
                        </button>
                        `
                    }

                    return "You may leave feedback after the workshop has completed"
                }
            }
        ]
    });

}