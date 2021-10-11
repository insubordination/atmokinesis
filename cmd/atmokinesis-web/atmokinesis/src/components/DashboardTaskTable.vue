<template>
  <div id="dashboard_task_table">
    <div class="tasks-container">
      <div class="tasks-container-title">
        <h1><span><svg xmlns="http://www.w3.org/2000/svg" width="30" height="30" fill="currentColor"
                       class="bi bi-list-task" viewBox="0 0 16 16"><path fill-rule="evenodd"
                                                                         d="M2 2.5a.5.5 0 0 0-.5.5v1a.5.5 0 0 0 .5.5h1a.5.5 0 0 0 .5-.5V3a.5.5 0 0 0-.5-.5H2zM3 3H2v1h1V3z"/><path
            d="M5 3.5a.5.5 0 0 1 .5-.5h9a.5.5 0 0 1 0 1h-9a.5.5 0 0 1-.5-.5zM5.5 7a.5.5 0 0 0 0 1h9a.5.5 0 0 0 0-1h-9zm0 4a.5.5 0 0 0 0 1h9a.5.5 0 0 0 0-1h-9z"/><path
            fill-rule="evenodd"
            d="M1.5 7a.5.5 0 0 1 .5-.5h1a.5.5 0 0 1 .5.5v1a.5.5 0 0 1-.5.5H2a.5.5 0 0 1-.5-.5V7zM2 7h1v1H2V7zm0 3.5a.5.5 0 0 0-.5.5v1a.5.5 0 0 0 .5.5h1a.5.5 0 0 0 .5-.5v-1a.5.5 0 0 0-.5-.5H2zm1 .5H2v1h1v-1z"/></svg>
        </span> Tasks
        </h1>
      </div>
      <div class="tasks-container-body">
        <table class="table">
          <thead>
          <tr>
            <th scope="col">Task</th>
            <th scope="col">Status</th>
            <th scope="col">Schedule</th>
            <th scope="col">Next Run</th>
            <th scope="col">History</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="task in tasks" :key="task.id" v-b-modal="'modal-task-' + task.id">
            <td><span style="font-weight: bold">{{ task.id }}</span></td>
            <td class="align-center" v-html="formatStatus(task.status)">
            </td>
            <td>{{ task.schedule }}</td>
            <td>{{ new Date(task.next_run) }}
              <div class="progress">
                <div class="progress-bar progress-bar-striped progress-bar-animated" role="progressbar"
                     :aria-valuenow="task.progress" aria-valuemin="0"
                     :aria-valuemax="100" :style="task.progressWidth">{{ task.remainingMinutes }}
                </div>
              </div>
            </td>
            <td>
              <div class="row" v-html="formatHistory(4, task.history)">
              </div>
            </td>
            <b-modal :id="'modal-task-' + task.id" size="xl" @shown="highlightSyntax" centered :title="task.id">
              <div class="card-body">
                <h5 class="card-title">Task Details</h5>
                <div class="row">
                  <div class="col-2"><p>Current Schedule</p></div>
                  <div class="col-10">
                                    <pre class="line-numbers"><code
                                        class="language-bash">{{ task.schedule }} / {{
                                        cronExplain(task.schedule)
                                      }}</code></pre>
                  </div>
                </div>
                <div class="row">
                  <div class="col-2">Next Run</div>
                  <div class="col">
                    <pre class="line-numbers"><code
                        class="language-bash">{{ new Date(task.next_run) }}</code></pre>
                  </div>
                </div>
                <div class="row">
                  <div class="col-2"></div>
                  <div class="col">
                    <div class="btn-group" role="group" aria-label="Basic example">
                      <button type="button" class="btn btn-success">Run Now</button>
                      <button type="button" class="btn btn-warning">Skip next run starting in
                        {{ task.remainingMinutes.replace("until start", "") }}
                      </button>
                      <button type="button" class="btn btn-danger">Pause</button>
                    </div>
                  </div>
                </div>
              </div>
              <div class="row d-flex justify-content-center mt-70 mb-70">
                <div class="col-lg-12">
                  <h5 class="card-title"> Event Timeline <span
                      style="font-weight: 100;font-style: italic">(last 20)</span></h5>
                  <div class="vertical-timeline vertical-timeline--animate vertical-timeline--one-column">
                    <div class="vertical-timeline-item vertical-timeline-element"
                         v-for="event in task.history.slice(Math.max(task.history.length - 20, 1))"
                         :key="event.ExecutionDate">
                      <div>
                        <div class="vertical-timeline-element-content bounce-in">
                          <h4 class="timeline-title">{{ event.status }}</h4>
                          <div class="row">
                            <div class="col-1"><p>logs</p></div>
                            <div class="col-11">
                                    <pre class="line-numbers" v-if="event.logs" style="max-height: 400px"><code
                                        class="language-json">{{ event.logs }}</code></pre>
                              <pre class="line-numbers" v-if="!event.logs"><code
                                  class="language-json">No Logs Sent</code></pre>
                            </div>
                          </div>
                          <span class="vertical-timeline-element-date">{{ event.execution_time }}</span>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </b-modal>
          </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
<script>
import * as Prism from 'prismjs'
import 'prismjs/components/prism-bash'
import 'prismjs/components/prism-javascript'
import 'prismjs/components/prism-json'
import 'prismjs/components/prism-liquid'
import 'prismjs/components/prism-markdown'
import 'prismjs/components/prism-markup-templating'
import 'prismjs/components/prism-php'
import 'prismjs/components/prism-scss'
import 'prismjs/themes/prism-okaidia.css'
import cronstrue from 'cronstrue';

export default {
  name: "DashboardTaskTable",
  data: () => ({
    tasks: [],
  }),
  mounted: function () {
    let connection = new WebSocket('ws://127.0.0.1:8082/taskstatus')
    connection.onopen = () => {
      connection.send("get");
    };
    connection.onmessage = ({data}) => {
      this.tasks = JSON.parse(data);
    };
  },
  created: function () {
    setInterval(() => {
      this.tasks.forEach((task, index) => {
        let progress = this.timerValue(task.last_run, task.next_run);
        task.progressWidth = progress.width;
        task.progress = progress.percent;
        task.remainingMinutes = progress.minutesLeft;
        if (task.progressWidth && task.progress) {
          this.$set(this.tasks, index, task)
        }
      })

    }, 1000);
  },
  methods: {
    cronExplain: (cron) => {
      return cronstrue.toString(cron);
    },
    prettifyJSON: (json) => {
      return JSON.stringify(JSON.parse(json), undefined, 4);
    },
    highlightSyntax: () => {
      Prism.highlightAll();
    },
    timerValue: (prevRun, nextRun) => {
      let now = new Date().getTime();
      let prev = Date.parse(prevRun);
      let nxt = Date.parse(nextRun);
      let diffFromNow = (nxt - now) / 1000;
      let totalDiff = (nxt - prev) / 1000;
      let width = "width: " + (100 - Math.round(diffFromNow / totalDiff * 100)) + "%"
      let percent = (Math.round(diffFromNow / totalDiff * 100))
      let minutesLeft = Math.floor(diffFromNow / 60)
      if (minutesLeft >= 1440) {
        minutesLeft = "less than " + Math.floor(minutesLeft / 120) + " day(s) until start"
      }
      if (minutesLeft > 60 && minutesLeft < 1440) {
        minutesLeft = "less than " + Math.floor(minutesLeft / 60) + " hr. until start"
      }
      if (minutesLeft <= 60 && minutesLeft !== 0) {
        minutesLeft = "less than " + Math.round(minutesLeft) + 1 + " min. until start"
      }
      if (minutesLeft === 0) {
        minutesLeft = "its the final countdown, " + Math.round(diffFromNow)
      }

      return {width, percent, minutesLeft};
    },
    formatHistory: (last, history) => {
      let historyString = '';
      let successTemplate = `<div class="col-1" style="color: #66ff69"><svg xmlns="http://www.w3.org/2000/svg" width="25" height="25" fill="currentColor" class="bi bi-check-circle-fill" viewBox="0 0 16 16">
                  <path d="M16 8A8 8 0 1 1 0 8a8 8 0 0 1 16 0zm-3.97-3.03a.75.75 0 0 0-1.08.022L7.477 9.417 5.384 7.323a.75.75 0 0 0-1.06 1.06L6.97 11.03a.75.75 0 0 0 1.079-.02l3.992-4.99a.75.75 0 0 0-.01-1.05z"/>
                </svg></div>`
      let failureTemplate = `<div class="col-1" style="color: #ff3434"><svg xmlns="http://www.w3.org/2000/svg" width="25" height="25" fill="currentColor" class="bi bi-x-circle" viewBox="0 0 16 16">
                  <path d="M8 15A7 7 0 1 1 8 1a7 7 0 0 1 0 14zm0 1A8 8 0 1 0 8 0a8 8 0 0 0 0 16z"/>
                  <path d="M4.646 4.646a.5.5 0 0 1 .708 0L8 7.293l2.646-2.647a.5.5 0 0 1 .708.708L8.707 8l2.647 2.646a.5.5 0 0 1-.708.708L8 8.707l-2.646 2.647a.5.5 0 0 1-.708-.708L7.293 8 4.646 5.354a.5.5 0 0 1 0-.708z"/>
                </svg></div>`
      Math.min(last, history.length)
      for (let i = 0; i < last; i++) {
        if (history[i].status === "Success") {
          historyString += successTemplate;
        }
        if (history[i].status === "Failing") {
          historyString += failureTemplate
        }
      }
      return historyString;
    },
    formatStatus: (status) => {
      switch (status) {
        case 'Pending Run':
          return `<svg xmlns="http://www.w3.org/2000/svg" width="25" height="25" fill="currentColor"
             className="bi bi-hourglass-split" viewBox="0 0 16 16">
          <path
              d="M2.5 15a.5.5 0 1 1 0-1h1v-1a4.5 4.5 0 0 1 2.557-4.06c.29-.139.443-.377.443-.59v-.7c0-.213-.154-.451-.443-.59A4.5 4.5 0 0 1 3.5 3V2h-1a.5.5 0 0 1 0-1h11a.5.5 0 0 1 0 1h-1v1a4.5 4.5 0 0 1-2.557 4.06c-.29.139-.443.377-.443.59v.7c0 .213.154.451.443.59A4.5 4.5 0 0 1 12.5 13v1h1a.5.5 0 0 1 0 1h-11zm2-13v1c0 .537.12 1.045.337 1.5h6.326c.216-.455.337-.963.337-1.5V2h-7zm3 6.35c0 .701-.478 1.236-1.011 1.492A3.5 3.5 0 0 0 4.5 13s.866-1.299 3-1.48V8.35zm1 0v3.17c2.134.181 3 1.48 3 1.48a3.5 3.5 0 0 0-1.989-3.158C8.978 9.586 8.5 9.052 8.5 8.351z"/>
        </svg>`
        case 'Running':
          return `<svg xmlns="http://www.w3.org/2000/svg" width="25" height="25" fill="currentColor" class="bi bi-wind" viewBox="0 0 16 16">
  <path d="M12.5 2A2.5 2.5 0 0 0 10 4.5a.5.5 0 0 1-1 0A3.5 3.5 0 1 1 12.5 8H.5a.5.5 0 0 1 0-1h12a2.5 2.5 0 0 0 0-5zm-7 1a1 1 0 0 0-1 1 .5.5 0 0 1-1 0 2 2 0 1 1 2 2h-5a.5.5 0 0 1 0-1h5a1 1 0 0 0 0-2zM0 9.5A.5.5 0 0 1 .5 9h10.042a3 3 0 1 1-3 3 .5.5 0 0 1 1 0 2 2 0 1 0 2-2H.5a.5.5 0 0 1-.5-.5z"/>
</svg>`
        case 'Stopped':
          return `<svg xmlns="http://www.w3.org/2000/svg" width="25" height="25" fill="currentColor"
               className="bi bi-cone-striped" viewBox="0 0 16 16">
            <path
                d="m9.97 4.88.953 3.811C10.159 8.878 9.14 9 8 9c-1.14 0-2.158-.122-2.923-.309L6.03 4.88C6.635 4.957 7.3 5 8 5s1.365-.043 1.97-.12zm-.245-.978L8.97.88C8.718-.13 7.282-.13 7.03.88L6.275 3.9C6.8 3.965 7.382 4 8 4c.618 0 1.2-.036 1.725-.098zm4.396 8.613a.5.5 0 0 1 .037.96l-6 2a.5.5 0 0 1-.316 0l-6-2a.5.5 0 0 1 .037-.96l2.391-.598.565-2.257c.862.212 1.964.339 3.165.339s2.303-.127 3.165-.339l.565 2.257 2.391.598z"/>
          </svg>`
        case 'Failing':
          return `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor"
                     className="bi bi-bug-fill" viewBox="0 0 16 16">
                  <path
                      d="M4.978.855a.5.5 0 1 0-.956.29l.41 1.352A4.985 4.985 0 0 0 3 6h10a4.985 4.985 0 0 0-1.432-3.503l.41-1.352a.5.5 0 1 0-.956-.29l-.291.956A4.978 4.978 0 0 0 8 1a4.979 4.979 0 0 0-2.731.811l-.29-.956z"/>
                  <path
                      d="M13 6v1H8.5v8.975A5 5 0 0 0 13 11h.5a.5.5 0 0 1 .5.5v.5a.5.5 0 1 0 1 0v-.5a1.5 1.5 0 0 0-1.5-1.5H13V9h1.5a.5.5 0 0 0 0-1H13V7h.5A1.5 1.5 0 0 0 15 5.5V5a.5.5 0 0 0-1 0v.5a.5.5 0 0 1-.5.5H13zm-5.5 9.975V7H3V6h-.5a.5.5 0 0 1-.5-.5V5a.5.5 0 0 0-1 0v.5A1.5 1.5 0 0 0 2.5 7H3v1H1.5a.5.5 0 0 0 0 1H3v1h-.5A1.5 1.5 0 0 0 1 11.5v.5a.5.5 0 1 0 1 0v-.5a.5.5 0 0 1 .5-.5H3a5 5 0 0 0 4.5 4.975z"/>
                </svg>`
      }
    }
  }
}
</script>

<style scoped>
thead {
  color: #f1f1f1;
  background-color: #464646;
}

td {
  color: #f1f1f1;
  background-color: #666666;
}

.progress-bar {
  color: black;
  font-weight: bold;
}

.tasks-container-body {
  background-color: #f6f6f6;
  border-radius: 0 0 2rem 2rem;
  border-width: 0;
  padding: 1rem 1rem 1rem 1rem;
  margin: 0 0 0 0;
  height: 100%;
  min-height: 6rem;
}

.tasks-container-title h1 {
  background-color: rgba(0, 223, 253, 0.83);
  border-width: 0;
  border-radius: 2rem 2rem 0 0;
  padding: 1rem 0 0 1rem;
  margin: 1rem 0 0 0;
  color: white;
  font-weight: 100;
  font-size: 2rem;
  height: 80px;
}

.table-label {
  font-weight: 100;
  font-size: 2rem;
  padding-top: 2rem;
  font-style: italic;
}

body {
  background-color: #eee
}

.mt-70 {
  margin-top: 70px
}

.mb-70 {
  margin-bottom: 70px
}

.card {
  box-shadow: 0 0.46875rem 2.1875rem rgba(4, 9, 20, 0.03), 0 0.9375rem 1.40625rem rgba(4, 9, 20, 0.03), 0 0.25rem 0.53125rem rgba(4, 9, 20, 0.05), 0 0.125rem 0.1875rem rgba(4, 9, 20, 0.03);
  border-width: 0;
  transition: all .2s
}

.card {
  position: relative;
  display: flex;
  flex-direction: column;
  min-width: 0;
  word-wrap: break-word;
  background-color: #fff;
  background-clip: border-box;
  border: 1px solid rgba(26, 54, 126, 0.125);
  border-radius: .25rem
}

.card-body {
  flex: 1 1 auto;
  padding: 1.25rem
}

.vertical-timeline {
  width: 100%;
  position: relative;
  padding: 1.5rem 0 1rem
}

.vertical-timeline::before {
  content: '';
  position: absolute;
  top: 0;
  left: 230px;
  height: 100%;
  width: 4px;
  background: #e9ecef;
  border-radius: .25rem
}

.vertical-timeline-element {
  position: relative;
  margin: 0 0 1rem
}

.vertical-timeline--animate .vertical-timeline-element-icon.bounce-in {
  visibility: visible;
  animation: cd-bounce-1 .8s
}

.vertical-timeline-element-icon {
  position: absolute;
  top: 0;
  left: 230px
}

.vertical-timeline-element-icon .badge-dot-xl {
  box-shadow: 0 0 0 5px #fff
}

.badge-dot-xl {
  width: 18px;
  height: 18px;
  position: relative
}

.badge-dot-xl::before {
  content: '';
  width: 10px;
  height: 10px;
  border-radius: .25rem;
  position: absolute;
  left: 50%;
  top: 50%;
  margin: -5px 0 0 -5px;
  background: #fff
}

.vertical-timeline-element-content {
  position: relative;
  margin-left: 250px;
  font-size: .8rem
}

.vertical-timeline-element-content .timeline-title {
  font-size: .8rem;
  text-transform: uppercase;
  margin: 0 0 .5rem;
  padding: 2px 0 0;
  font-weight: bold
}

.vertical-timeline-element-content .vertical-timeline-element-date {
  display: block;
  position: absolute;
  left: -230px;
  top: 0;
  padding-right: 50px;
  text-align: right;
  color: #adb5bd;
  font-size: .7619rem;
  white-space: nowrap
}

.vertical-timeline-element-content:after {
  content: "";
  display: table;
  clear: both
}
</style>