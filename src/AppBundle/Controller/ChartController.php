<?php

namespace AppBundle\Controller;

use AppBundle\Repository\LogRepository;
use Doctrine\DBAL\Exception;
use Doctrine\ORM\EntityManagerInterface;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\Cache\Adapter\AdapterInterface;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;

class ChartController extends AbstractController
{
    /**
     * @var array $data
     */
    protected $data;

    /**
     * @var LogRepository $repo
     */
    private $repo;

    /**
     * @var EntityManagerInterface $entity
     */
    private $entity;

    /**
     * @var AdapterInterface $cache
     */
    private $cache;

    public function __construct(EntityManagerInterface $entity, LogRepository $repo, AdapterInterface $cache)
    {
        $this->entity = $entity;
        $this->repo = $repo;
        $this->cache = $cache;
    }

    /**
     * @Route("/chart/{year}/{month}")
     * @throws Exception|\Doctrine\DBAL\Driver\Exception
     * @throws \Psr\Cache\InvalidArgumentException
     */
    public function chartAction(Request $request, int $year, string $month): Response
    {
        $sql = $this->repo->getContent('Requests.sql');

//        $logs = $entityManager
//            ->getConnection()
//            ->executeQuery($sql)
//            ->fetchAllAssociative();

        $stmt = $this->entity
            ->getConnection()
            ->prepare($sql);
        //$stmt->bindValue('request', '/VigieTextofPZN.php');
        //$stmt->bindValue('request', '/schnittstellen/webservices/index.php');
        $stmt->bindValue('request', $request->query->get('request'));
        $stmt->bindValue('date', "$year-$month-%");

        $key = md5($sql . $year . $month . $request->query->get('request'));

        $item = $this->cache->getItem($key);

        if ($item->isHit()) {
            $logs = $item->get();
        } else {
            $stmt->executeQuery();
            $logs = $stmt->fetchAll();
            $item->set($logs);
            $this->cache->save($item);
        }

        foreach ($logs as $log) {
            $date = date_create($log['Date']);
            $this->data['labels'][] = date_format($date,'d');
            $this->data['number'][] = $log['Number'];
            $this->data['median'][] = round($log['Average']);
        }

        return $this->render('@App/chart/index.html.twig', [
            'labels' => json_encode($this->data['labels']),
            'number' => json_encode($this->data['number']),
            'median' => json_encode($this->data['median'])
        ]);

    }
}

