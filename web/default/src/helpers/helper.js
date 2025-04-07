import {CHANNEL_OPTIONS} from '../constants';

let channelMap = undefined;

export function getChannelOption(channelId) {
    if (channelMap === undefined) {
        channelMap = {};
        CHANNEL_OPTIONS.forEach((option) => {
            channelMap[option.key] = option;
        });
    }
    return channelMap[channelId];
}



  
  export function timestamp2string1(timestamp, dataExportDefaultTime = 'hour') {
    let date = new Date(timestamp * 1000);
    // let year = date.getFullYear().toString();
    let month = (date.getMonth() + 1).toString();
    let day = date.getDate().toString();
    let hour = date.getHours().toString();
    if (day === '24') {
      console.log("timestamp", timestamp);
    }
    if (month.length === 1) {
      month = '0' + month;
    }
    if (day.length === 1) {
      day = '0' + day;
    }
    if (hour.length === 1) {
      hour = '0' + hour;
    }
    let str = month + '-' + day;
    if (dataExportDefaultTime === 'hour') {
      str += ' ' + hour + ':00';
    } else if (dataExportDefaultTime === 'week') {
      let nextWeek = new Date(timestamp * 1000 + 6 * 24 * 60 * 60 * 1000);
      let nextMonth = (nextWeek.getMonth() + 1).toString();
      let nextDay = nextWeek.getDate().toString();
      if (nextMonth.length === 1) {
        nextMonth = '0' + nextMonth;
      }
      if (nextDay.length === 1) {
        nextDay = '0' + nextDay;
      }
      str += ' - ' + nextMonth + '-' + nextDay;
    }
    return str;
  }

