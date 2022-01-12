select strftime(:view, date) as datetime,
       method,
       request,
       param,
       port,
       status,
       count(*)      as number,
       avg(duration) as average
from Log
where request = :request
  and param = :param
  and status = :status
  and date like :date
group by datetime, method, request, param, port, status
order by datetime, number desc