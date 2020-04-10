<?php
/**
 * @see       https://github.com/zendframework/zend-expressive-helpers for the canonical source repository
 * @copyright Copyright (c) 2015-2017 Zend Technologies USA Inc. (https://www.zend.com)
 * @license   https://github.com/zendframework/zend-expressive-helpers/blob/master/LICENSE.md New BSD License
 */

declare(strict_types=1);

namespace Zend\Expressive\Helper\Exception;

use DomainException;
use Psr\Container\ContainerExceptionInterface;

class MissingHelperException extends DomainException implements
    ContainerExceptionInterface,
    ExceptionInterface
{
}
