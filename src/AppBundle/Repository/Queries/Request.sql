select strftime('%Y-%m', date) as datetime,
       method,
       request,
       param,
       port,
       status,
       avg(duration)           as duration,
       count(*)                as number
from Log
where (request like :search
    or param like :search)
group by datetime, method, request, param, port, status
order by datetime desc, number desc
limit 200