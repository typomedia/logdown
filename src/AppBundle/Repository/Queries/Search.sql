select strftime(:view, date) as datetime,
       method,
       request,
       param,
       port,
       status,
       avg(duration)              as duration,
       count(*)                   as number
from Log
where request = :request
  and param = :param
  and date like :date
group by datetime, method, request, param, port, status
order by datetime desc, number desc